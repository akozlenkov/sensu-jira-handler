# sensu-jira-handler-go-jira-v2
This is just a fork of akozlenkov's original work to use go-jira-v2.

```
type: Handler
api_version: core/v2
metadata:
  name: jira
  namespace: default
spec:
  command: /gobins/sensu-jira-handler --jira-url 'https://jirainstance.atlassian.net' --jira-project 'PRJ' --jira-issue-type 'Task'
  env_vars: null
  filters: null
  handlers: null
  runtime_assets: null
  secrets: null
  timeout: 10
  type: pipe
```