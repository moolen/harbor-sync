# Usage examples

## Map projects by name

Map harbor project to several namespaces. This will create a robot account in `my-project` **harbor project** and sync the credentials into `team-a` and `team-b`'s namespace as secret `central-project-token`.

```yml
spec:
  projectSelector:
  - type: Regex
    name: "my-project" # <--- specify harbor project
    robotAccountSuffix: "k8s-sync-robot"
    mapping:
    - type: Translate
      namespace: "team-a" # <--- target namespace
      secret: "my-project-pull-token" # <--- target secret name
    - type: Translate
      namespace: "team-b"
      secret: "my-project-pull-token"
```



## Map projects using a regular expression

You can specify regular expressions to map a **large number** of projects to namespaces. This maps harbor teams with the prefix `team-`. E.g. Harbor project `team-frontend` maps to k8s namespace `team-frontend`. The secret's name will always be `my-pull-token`. Non-existent k8s namespaces will be ignored.

```yaml
spec:
  projectSelector:
  - type: Regex
    name: "team-(.*)" # find harbor projects matching this expression
    robotAccountSuffix: "k8s-sync-robot"
    mapping:
    - type: Translate
      namespace: "team-$1"    # references capturing group from the above projectSelector.name
      secret: "team-$1-pull-token" # also here
```



## Map projects using regular expressions #2

You have **one** harbor project and want to deploy the pull secrets **into several namespaces** matching a regular expression. E.g. pull tokens for the `platform-team` project should be distributed into all namespaces matching `team-.*`.

Use a `type: Match` on a mapping to say: hey, find namespaces using this **regular expression** at the namespace field rather than re-using the project name using `type: Translate`.


```yaml
spec:
  projectSelector:
  - type: Regex
    name: "platform-team"
    robotAccountSuffix: "k8s-sync-robot"
    mapping:
    - type: Match  # treat namespace as regexp
      namespace: "team-.*" # if ns matches this it will receive the secret
      secret: "platform-pull-token" # you can still use the capturing group from projectSelector.Name here
```

## Mapping Projects

A `mapping` defines how to lookup namespaces in the cluster. Generally there are two lookup types: `Translate` and `Match`.

### Translate
**Translate** will take the Harbor project name into account when looking up namespaces. The `ProjectSelector.ProjectName` can be a regular expression which holds capturing groups. The idea is to inject those capturing groups when finding namespaces.

Example:

Harbor: we have two projects, `team-frontend` and `team-backend`. We select them using `team-(.*)` in the `ProjectSelector.ProjectName`. And map them to kubernetes namespaces `squad-$1`. The `$1` will be replaced with `frontend` and `backend` respectively. In the end each namespaces will have **only it's own** secret: `team-frontend` will only have the secret of Harbor project `team-frontend`. Namespace `team-backend` will only have the secret of Harbor project `team-backend`.

### Match
**Match** doesn't care about the `ProjectSelector.ProjectName`. It will just find **namespaces** in the cluster that match the **regular expression**.

Example 1:

Harbor: we have one project, `team-platform`. By setting the field `ProjectMapping.Namespace` to `team-.*` we deploy the robot account secret to namespaces

Example 2:

Harbor: we have two projects, `team-platform` and `team-operations`. By setting `ProjectMapping.Namespace` to `team-.*` we deploy the robot accounts of both the `platform` and `operations` project into the namespace. To avoid naming conflicts on the secrets we set `ProjectMapping.Secret` to `$1-pull-token`. The result is: All namespaces matching `team-.*` will have the secrets `platform-pull-token` and `operations-pull-token`.
