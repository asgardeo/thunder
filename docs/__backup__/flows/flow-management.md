# Flow Management

This guide covers updating, versioning, and deleting flows.

## Updating Flows

When you update a flow, a new version is created automatically. The previous version is retained for history.

```bash
curl -X PUT https://localhost:8090/flows/{flowId} \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <token>' \
  -d '{
    "name": "Updated Flow Name",
    "nodes": [...]
  }'
```

> [!NOTE]
> The `handle` property cannot be changed after creation.

---

## Version History & Restoration

### View Version History

```bash
curl -X GET https://localhost:8090/flows/{flowId}/versions \
  -H 'Authorization: Bearer <token>'
```

### Get Specific Version

```bash
curl -X GET https://localhost:8090/flows/{flowId}/versions/{version} \
  -H 'Authorization: Bearer <token>'
```

### Restore Previous Version

```bash
curl -X POST https://localhost:8090/flows/{flowId}/restore \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <token>' \
  -d '{
    "version": 3
  }'
```

> [!NOTE]
> The number of versions retained is configurable via the server configuration. Older versions beyond the `max_version_history` limit are automatically removed when new versions are created.
> 
> ```yaml
> flow:
>   max_version_history: 10
> ```

---

## Deleting Flows

```bash
curl -X DELETE https://localhost:8090/flows/{flowId} \
  -H 'Authorization: Bearer <token>'
```

---

## API Reference

See the [Flow Management API](/api/flow-management.yaml) for the complete API specification.
