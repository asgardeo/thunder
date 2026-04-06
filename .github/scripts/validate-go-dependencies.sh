#!/bin/bash

set -e

BASE_SHA=$1
HEAD_SHA=$2
APPROVED_LIST=$3

echo "======================================"
echo "Go Dependency Validation"
echo "======================================"
echo "Base SHA: $BASE_SHA"
echo "Head SHA: $HEAD_SHA"
echo "Approved list: $APPROVED_LIST"
echo ""

# Install yq for YAML parsing
echo "Installing yq..."
wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
chmod +x /usr/local/bin/yq

# Function to parse version constraint
check_version_constraint() {
    local module=$1
    local version=$2
    local constraint=$3

    # Handle pseudo versions
    if [[ "$constraint" == "pseudo" ]]; then
        if [[ "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+-[0-9]{14}-[a-f0-9]{12}$ ]]; then
            return 0
        else
            return 1
        fi
    fi

    # Handle wildcard (all versions allowed)
    if [[ "$constraint" == "*" ]]; then
        return 0
    fi

    # Remove 'v' prefix for comparison
    version_clean="${version#v}"

    # Handle range constraints (e.g., ">=v1.7.0 <v1.9.1")
    if [[ "$constraint" =~ \  ]]; then
        # Split into individual constraints
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

    # Extract operator and constraint version
    if [[ "$constraint" =~ ^(\>\=|\<\=|\>|\<|=)v?(.+)$ ]]; then
        operator="${BASH_REMATCH[1]}"
        constraint_version="${BASH_REMATCH[2]}"
    else
        echo "Invalid constraint format: $constraint" >&2
        return 1
    fi

    # Use sort -V for version comparison
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

# Find all go.mod files that changed
CHANGED_GO_MODS=$(git diff --name-only $BASE_SHA $HEAD_SHA | grep 'go.mod$' || true)

if [ -z "$CHANGED_GO_MODS" ]; then
    echo "No go.mod files were modified"
    exit 0
fi

echo "Changed go.mod files:"
echo "$CHANGED_GO_MODS"
echo ""

# Initialize result files
> /tmp/new_deps.txt
> /tmp/updated_deps.txt
> /tmp/unapproved_deps.txt
> /tmp/validated_deps.txt

# Process each changed go.mod file
for GO_MOD in $CHANGED_GO_MODS; do
    echo "Processing: $GO_MOD"
    echo "----------------------------------------"

    # Get the old and new versions of go.mod
    git show $BASE_SHA:$GO_MOD > /tmp/go.mod.old 2>/dev/null || touch /tmp/go.mod.old
    git show $HEAD_SHA:$GO_MOD > /tmp/go.mod.new 2>/dev/null || continue

    # Extract direct dependencies (not indirect)
    grep -E '^\s+[a-zA-Z0-9]' /tmp/go.mod.old | grep -v '// indirect' | awk '{print $1, $2}' | sort > /tmp/deps.old
    grep -E '^\s+[a-zA-Z0-9]' /tmp/go.mod.new | grep -v '// indirect' | awk '{print $1, $2}' | sort > /tmp/deps.new

    # Find new dependencies
    NEW_DEPS=$(comm -13 /tmp/deps.old /tmp/deps.new | awk '{print $1}')

    # Find updated dependencies (same module, different version)
    UPDATED_DEPS=$(comm -12 <(awk '{print $1}' /tmp/deps.old | sort) <(awk '{print $1}' /tmp/deps.new | sort))

    # Check new dependencies
    for MODULE in $NEW_DEPS; do
        VERSION=$(grep "^$MODULE " /tmp/deps.new | awk '{print $2}')
        echo "NEW: $MODULE $VERSION" | tee -a /tmp/new_deps.txt

        # Check if module is approved
        APPROVED_VERSIONS=$(yq eval ".dependencies[] | select(.module == \"$MODULE\") | .versions[].version" "$APPROVED_LIST" 2>/dev/null || echo "")

        if [ -z "$APPROVED_VERSIONS" ]; then
            echo "  ❌ NOT APPROVED: Module not in approved list" | tee -a /tmp/unapproved_deps.txt
            echo "$MODULE $VERSION - Module not in approved list" >> /tmp/unapproved_deps.txt
        else
            # Check if version matches any constraint
            MATCH_FOUND=false
            while IFS= read -r CONSTRAINT; do
                if check_version_constraint "$MODULE" "$VERSION" "$CONSTRAINT"; then
                    MATCH_FOUND=true
                    echo "  ✅ APPROVED: Matches constraint $CONSTRAINT"
                    echo "$MODULE $VERSION - Approved (constraint: $CONSTRAINT)" >> /tmp/validated_deps.txt
                    break
                fi
            done <<< "$APPROVED_VERSIONS"

            if [ "$MATCH_FOUND" = false ]; then
                echo "  ❌ NOT APPROVED: Version $VERSION does not match any approved constraint"
                echo "$MODULE $VERSION - Version does not match approved constraints: $APPROVED_VERSIONS" >> /tmp/unapproved_deps.txt
            fi
        fi
    done

    # Check updated dependencies
    for MODULE in $UPDATED_DEPS; do
        OLD_VERSION=$(grep "^$MODULE " /tmp/deps.old | awk '{print $2}')
        NEW_VERSION=$(grep "^$MODULE " /tmp/deps.new | awk '{print $2}')

        if [ "$OLD_VERSION" != "$NEW_VERSION" ]; then
            echo "UPDATED: $MODULE $OLD_VERSION -> $NEW_VERSION" | tee -a /tmp/updated_deps.txt

            # Check if new version is approved
            APPROVED_VERSIONS=$(yq eval ".dependencies[] | select(.module == \"$MODULE\") | .versions[].version" "$APPROVED_LIST" 2>/dev/null || echo "")

            if [ -z "$APPROVED_VERSIONS" ]; then
                echo "  ❌ NOT APPROVED: Module not in approved list"
                echo "$MODULE $NEW_VERSION - Module not in approved list" >> /tmp/unapproved_deps.txt
            else
                # Check if version matches any constraint
                MATCH_FOUND=false
                while IFS= read -r CONSTRAINT; do
                    if check_version_constraint "$MODULE" "$NEW_VERSION" "$CONSTRAINT"; then
                        MATCH_FOUND=true
                        echo "  ✅ APPROVED: Matches constraint $CONSTRAINT"
                        echo "$MODULE $OLD_VERSION -> $NEW_VERSION - Approved (constraint: $CONSTRAINT)" >> /tmp/validated_deps.txt
                        break
                    fi
                done <<< "$APPROVED_VERSIONS"

                if [ "$MATCH_FOUND" = false ]; then
                    echo "  ❌ NOT APPROVED: Version $NEW_VERSION does not match any approved constraint"
                    echo "$MODULE $NEW_VERSION - Version does not match approved constraints: $APPROVED_VERSIONS" >> /tmp/unapproved_deps.txt
                fi
            fi
        fi
    done

    echo ""
done

# Summary
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
