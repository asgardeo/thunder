#!/bin/bash

set -e

BASE_SHA=$1
HEAD_SHA=$2
APPROVED_LIST=$3
# Update scope value when using in other repositories.
REQUIRED_SCOPE="${REQUIRED_SCOPE:-thunder-id}"

if [ ! -s "$APPROVED_LIST" ]; then
  echo "Approved dependency registry not found or empty: $APPROVED_LIST" >&2
  exit 2
fi
if ! yq eval '.dependencies | length' "$APPROVED_LIST" >/dev/null 2>&1; then
  echo "Approved dependency registry is not valid YAML/schema: $APPROVED_LIST" >&2
  exit 2
fi

echo "======================================"
echo "Go Dependency Validation"
echo "======================================"
echo "Base SHA: $BASE_SHA"
echo "Head SHA: $HEAD_SHA"
echo "Approved list: $APPROVED_LIST"
echo "Required scope: $REQUIRED_SCOPE"
echo ""

echo "Installing yq..."
wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
chmod +x /usr/local/bin/yq

check_version_constraint() {
    local module=$1
    local version=$2
    local constraint=$3

    if [[ "$constraint" == "pseudo" ]]; then
        if [[ "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+-[0-9]{14}-[a-f0-9]{12}$ || \
              "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+-(0|[0-9A-Za-z-]+\.0)\.[0-9]{14}-[a-f0-9]{12}$ ]]; then
            return 0
        else
            return 1
        fi
    fi

    if [[ "$constraint" == "*" ]]; then
        return 0
    fi

    version_clean="${version#v}"

    if [[ "$constraint" =~ \  ]]; then
        IFS=' ' read -ra CONSTRAINTS <<< "$constraint"
        for c in "${CONSTRAINTS[@]}"; do
            if ! check_single_constraint "$version_clean" "$c"; then
                return 1
            fi
        done
        return 0
    else
        check_single_constraint "$version_clean" "$constraint"
    fi
}

check_single_constraint() {
    local version=$1
    local constraint=$2

    if [[ "$constraint" =~ ^(\>\=|\<\=|\>|\<|=)v?(.+)$ ]]; then
        operator="${BASH_REMATCH[1]}"
        constraint_version="${BASH_REMATCH[2]}"
    else
        echo "Invalid constraint format: $constraint" >&2
        return 1
    fi

    case "$operator" in
        ">=")
            [ "$(printf '%s\n' "$constraint_version" "$version" | sort -V | head -n1)" = "$constraint_version" ]
            ;;
        "<=")
            [ "$(printf '%s\n' "$version" "$constraint_version" | sort -V | head -n1)" = "$version" ]
            ;;
        ">")
            [ "$(printf '%s\n' "$constraint_version" "$version" | sort -V | head -n1)" = "$constraint_version" ] && \
            [ "$version" != "$constraint_version" ]
            ;;
        "<")
            [ "$(printf '%s\n' "$version" "$constraint_version" | sort -V | head -n1)" = "$version" ] && \
            [ "$version" != "$constraint_version" ]
            ;;
        "=")
            [ "$version" = "$constraint_version" ]
            ;;
        *)
            echo "Unknown operator: $operator" >&2
            return 1
            ;;
    esac
}

CHANGED_GO_MODS=$(git diff --name-only $BASE_SHA $HEAD_SHA | grep 'go.mod$' || true)

if [ -z "$CHANGED_GO_MODS" ]; then
    echo "No go.mod files were modified"
    exit 0
fi

echo "Changed go.mod files:"
echo "$CHANGED_GO_MODS"
echo ""

> /tmp/new_deps.txt
> /tmp/updated_deps.txt
> /tmp/unapproved_deps.txt
> /tmp/validated_deps.txt
> /tmp/all_deps_status.txt

for GO_MOD in $CHANGED_GO_MODS; do
    echo "Processing: $GO_MOD"
    echo "----------------------------------------"

    git show $BASE_SHA:$GO_MOD > /tmp/go.mod.old 2>/dev/null || touch /tmp/go.mod.old
    git show $HEAD_SHA:$GO_MOD > /tmp/go.mod.new 2>/dev/null || continue

    awk '
      $1=="require" && $2=="(" { in_require=1; next }
      in_require && $1==")"   { in_require=0; next }
      $1=="require" && $2!="(" && $0 !~ /\/\/ indirect/ { print $2, $3; next }
      in_require && $0 !~ /\/\/ indirect/ { print $1, $2 }
    ' /tmp/go.mod.old | sort > /tmp/deps.old

    awk '
      $1=="require" && $2=="(" { in_require=1; next }
      in_require && $1==")"   { in_require=0; next }
      $1=="require" && $2!="(" && $0 !~ /\/\/ indirect/ { print $2, $3; next }
      in_require && $0 !~ /\/\/ indirect/ { print $1, $2 }
    ' /tmp/go.mod.new | sort > /tmp/deps.new

    NEW_DEPS=$(comm -13 \
      <(awk '{print $1}' /tmp/deps.old | sort -u) \
      <(awk '{print $1}' /tmp/deps.new | sort -u))

    UPDATED_DEPS=$(comm -12 \
      <(awk '{print $1}' /tmp/deps.old | sort -u) \
      <(awk '{print $1}' /tmp/deps.new | sort -u))

    for MODULE in $NEW_DEPS; do
        VERSION=$(awk -v module="$MODULE" '$1==module {print $2; exit}' /tmp/deps.new)
        echo "NEW: $MODULE $VERSION" | tee -a /tmp/new_deps.txt

        APPROVED_VERSIONS=$(yq eval ".dependencies[] | select(.module == \"$MODULE\") | .versions[].version" "$APPROVED_LIST" 2>/dev/null || echo "")

        if [ -z "$APPROVED_VERSIONS" ]; then
            echo "  ❌ NOT APPROVED: Module not in approved list" | tee -a /tmp/unapproved_deps.txt
            echo "$MODULE $VERSION - Module not in approved list" >> /tmp/unapproved_deps.txt
            echo "UNAPPROVED|$MODULE|$VERSION|Module not found in dependency registry" >> /tmp/all_deps_status.txt
        else
            MATCH_FOUND=false
            SCOPE_VALID=false
            MATCHED_CONSTRAINT=""

            VERSION_COUNT=$(yq eval ".dependencies[] | select(.module == \"$MODULE\") | .versions | length" "$APPROVED_LIST" 2>/dev/null || echo "0")

            for ((i=0; i<VERSION_COUNT; i++)); do
                CONSTRAINT=$(yq eval ".dependencies[] | select(.module == \"$MODULE\") | .versions[$i].version" "$APPROVED_LIST" 2>/dev/null)

                if check_version_constraint "$MODULE" "$VERSION" "$CONSTRAINT"; then
                    ALLOWED_SCOPES=$(yq eval ".dependencies[] | select(.module == \"$MODULE\") | .versions[$i].allowed_scopes[]" "$APPROVED_LIST" 2>/dev/null || echo "")

                    if printf '%s\n' "$ALLOWED_SCOPES" | grep -Fxq '*' || \
                       printf '%s\n' "$ALLOWED_SCOPES" | grep -Fxq "$REQUIRED_SCOPE"; then
                        MATCH_FOUND=true
                        SCOPE_VALID=true
                        MATCHED_CONSTRAINT="$CONSTRAINT"
                        echo "  ✅ APPROVED: Matches constraint $CONSTRAINT with valid scope"
                        echo "$MODULE $VERSION - Approved (constraint: $CONSTRAINT)" >> /tmp/validated_deps.txt
                        echo "APPROVED|$MODULE|$VERSION|Approved (constraint: $CONSTRAINT)" >> /tmp/all_deps_status.txt
                        break
                    else
                        MATCH_FOUND=true
                        SCOPE_VALID=false
                        SCOPES_LIST=$(echo "$ALLOWED_SCOPES" | tr '\n' ',' | sed 's/,$//' | sed 's/,/, /g')
                        echo "  ❌ SCOPE MISMATCH: Version matches but scope restriction does not allow $REQUIRED_SCOPE usage"
                        echo "$MODULE $VERSION - Scope mismatch (allowed scopes: $ALLOWED_SCOPES)" >> /tmp/unapproved_deps.txt
                        echo "UNAPPROVED|$MODULE|$VERSION|Scope mismatch - allowed scopes: [$SCOPES_LIST], but '$REQUIRED_SCOPE' scope is required" >> /tmp/all_deps_status.txt
                        break
                    fi
                fi
            done

            if [ "$MATCH_FOUND" = false ]; then
                CONSTRAINTS_LIST=$(echo "$APPROVED_VERSIONS" | tr '\n' ',' | sed 's/,$//' | sed 's/,/, /g')
                echo "  ❌ NOT APPROVED: Version $VERSION does not match any approved constraint"
                echo "$MODULE $VERSION - Version does not match approved constraints: $APPROVED_VERSIONS" >> /tmp/unapproved_deps.txt
                echo "UNAPPROVED|$MODULE|$VERSION|Version constraint not met - approved versions: [$CONSTRAINTS_LIST]" >> /tmp/all_deps_status.txt
            fi
        fi
    done

    for MODULE in $UPDATED_DEPS; do
        OLD_VERSION=$(awk -v module="$MODULE" '$1==module {print $2; exit}' /tmp/deps.old)
        NEW_VERSION=$(awk -v module="$MODULE" '$1==module {print $2; exit}' /tmp/deps.new)

        if [ "$OLD_VERSION" != "$NEW_VERSION" ]; then
            echo "UPDATED: $MODULE $OLD_VERSION -> $NEW_VERSION" | tee -a /tmp/updated_deps.txt

            APPROVED_VERSIONS=$(yq eval ".dependencies[] | select(.module == \"$MODULE\") | .versions[].version" "$APPROVED_LIST" 2>/dev/null || echo "")

            if [ -z "$APPROVED_VERSIONS" ]; then
                echo "  ❌ NOT APPROVED: Module not in approved list"
                echo "$MODULE $NEW_VERSION - Module not in approved list" >> /tmp/unapproved_deps.txt
                echo "UNAPPROVED|$MODULE|$NEW_VERSION (was $OLD_VERSION)|Module not found in dependency registry" >> /tmp/all_deps_status.txt
            else
                MATCH_FOUND=false
                SCOPE_VALID=false
                MATCHED_CONSTRAINT=""

                VERSION_COUNT=$(yq eval ".dependencies[] | select(.module == \"$MODULE\") | .versions | length" "$APPROVED_LIST" 2>/dev/null || echo "0")

                for ((i=0; i<VERSION_COUNT; i++)); do
                    CONSTRAINT=$(yq eval ".dependencies[] | select(.module == \"$MODULE\") | .versions[$i].version" "$APPROVED_LIST" 2>/dev/null)

                    if check_version_constraint "$MODULE" "$NEW_VERSION" "$CONSTRAINT"; then
                        ALLOWED_SCOPES=$(yq eval ".dependencies[] | select(.module == \"$MODULE\") | .versions[$i].allowed_scopes[]" "$APPROVED_LIST" 2>/dev/null || echo "")

                        if printf '%s\n' "$ALLOWED_SCOPES" | grep -Fxq '*' || \
                           printf '%s\n' "$ALLOWED_SCOPES" | grep -Fxq "$REQUIRED_SCOPE"; then
                            MATCH_FOUND=true
                            SCOPE_VALID=true
                            MATCHED_CONSTRAINT="$CONSTRAINT"
                            echo "  ✅ APPROVED: Matches constraint $CONSTRAINT with valid scope"
                            echo "$MODULE $OLD_VERSION -> $NEW_VERSION - Approved (constraint: $CONSTRAINT)" >> /tmp/validated_deps.txt
                            echo "APPROVED|$MODULE|$NEW_VERSION (was $OLD_VERSION)|Approved (constraint: $CONSTRAINT)" >> /tmp/all_deps_status.txt
                            break
                        else
                            MATCH_FOUND=true
                            SCOPE_VALID=false
                            SCOPES_LIST=$(echo "$ALLOWED_SCOPES" | tr '\n' ',' | sed 's/,$//' | sed 's/,/, /g')
                            echo "  ❌ SCOPE MISMATCH: Version matches but scope restriction does not allow $REQUIRED_SCOPE usage"
                            echo "$MODULE $NEW_VERSION - Scope mismatch (allowed scopes: $ALLOWED_SCOPES)" >> /tmp/unapproved_deps.txt
                            echo "UNAPPROVED|$MODULE|$NEW_VERSION (was $OLD_VERSION)|Scope mismatch - allowed scopes: [$SCOPES_LIST], but '$REQUIRED_SCOPE' scope is required" >> /tmp/all_deps_status.txt
                            break
                        fi
                    fi
                done

                if [ "$MATCH_FOUND" = false ]; then
                    CONSTRAINTS_LIST=$(echo "$APPROVED_VERSIONS" | tr '\n' ',' | sed 's/,$//' | sed 's/,/, /g')
                    echo "  ❌ NOT APPROVED: Version $NEW_VERSION does not match any approved constraint"
                    echo "$MODULE $NEW_VERSION - Version does not match approved constraints: $APPROVED_VERSIONS" >> /tmp/unapproved_deps.txt
                    echo "UNAPPROVED|$MODULE|$NEW_VERSION (was $OLD_VERSION)|Version constraint not met - approved versions: [$CONSTRAINTS_LIST]" >> /tmp/all_deps_status.txt
                fi
            fi
        fi
    done

    echo ""
done

echo "======================================"
echo "Validation Summary"
echo "======================================"

NEW_COUNT=$(wc -l < /tmp/new_deps.txt | tr -d ' ')
UPDATED_COUNT=$(wc -l < /tmp/updated_deps.txt | tr -d ' ')
UNAPPROVED_COUNT=$(wc -l < /tmp/unapproved_deps.txt | tr -d ' ')

echo "New dependencies: $NEW_COUNT"
echo "Updated dependencies: $UPDATED_COUNT"
echo "Unapproved dependencies: $UNAPPROVED_COUNT"
echo ""

if [ "$UNAPPROVED_COUNT" -gt 0 ]; then
    echo "❌ VALIDATION FAILED"
    echo ""
    echo "Unapproved dependencies found:"
    cat /tmp/unapproved_deps.txt
    exit 1
else
    echo "✅ VALIDATION PASSED"
    echo "All dependencies are approved!"
    exit 0
fi
