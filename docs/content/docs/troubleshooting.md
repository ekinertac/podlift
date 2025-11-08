---
title: Troubleshooting
weight: 60
---

### Error: Service name already in use

```
ERROR: Service 'myapp' is already deployed on 192.168.1.10

Existing deployment:
  Version: abc123
  Deployed: 1 hour ago
  Containers: 2

If this is the same application:
  This is normal - proceeding will update the existing deployment

If this is a DIFFERENT application:
  Change the service name in podlift.yml to avoid conflicts
```

**Cause**: You're trying to deploy a different app with the same service name.

**Scenario 1 - Same App (Redeployment):**

This is normal! You're updating your existing app. Proceed:
```bash
podlift deploy
```

**Scenario 2 - Different App (Conflict):**

You have two different apps trying to use the same service name.

Example:
```yaml
# blog-repo/podlift.yml
service: myapp  # ← Problem!

# api-repo/podlift.yml  
service: myapp  # ← Same name!
```

**Solution:** Use unique service names:

```yaml
# blog-repo/podlift.yml
service: blog  # ← Unique

# api-repo/podlift.yml
service: api   # ← Unique
```

Service names must be unique per server.

**Verification:**

Check what services are deployed:
```bash
ssh root@192.168.1.10 'docker ps --filter "label=podlift.service" --format "{{.Label \"podlift.service\"}}"'
```

Lists all podlift-managed services on that server.

---
