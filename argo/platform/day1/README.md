# Day 1 Baseline - ArgoCD Project Boundaries + RBAC

This Day 1 baseline sets up:
- Project boundaries with three AppProjects (`platform-core`, `team-dev`, `team-prod`)
- Group-to-role mappings through `argocd-rbac-cm`
- A parent app (`platform-root`) that bootstraps child apps for projects and RBAC

## Files

- `projects/appprojects.yaml`: AppProject boundaries and project roles
- `rbac/argocd-rbac-cm.yaml`: Global and project-scoped RBAC group mappings
- `apps/root-app.yaml`: Parent app-of-apps for platform bootstrap
- `apps/children/project-bootstrap.yaml`: Child app syncing `projects/`
- `apps/children/rbac-bootstrap.yaml`: Child app syncing `rbac/`

## Apply Order

Apply in this order the first time:

```bash
kubectl apply -f argo/platform/day1/projects/appprojects.yaml
kubectl apply -f argo/platform/day1/rbac/argocd-rbac-cm.yaml
kubectl apply -f argo/platform/day1/apps/root-app.yaml
```

Then verify:

```bash
argocd proj list
argocd app list
kubectl get cm argocd-rbac-cm -n argocd
```

## Customize Before Production

- Update all `repoURL` values if your Git remote changes.
- Replace group names (`dev-admins`, `prod-releasers`, etc.) to match your IdP groups.
- Tune `team-prod` sync windows for your release policy.

## Success Criteria

- Team apps cannot deploy outside allowed namespaces.
- `dev` users can sync only `team-dev` apps.
- `prod` release users can sync only `team-prod` apps.
- `platform-admins` can manage bootstrap and platform-level apps.
