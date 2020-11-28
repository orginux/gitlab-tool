# gitlab-tool
## Util for interact with GitLab for use in CI or in command line.
### Examples
#### Download artifacts
If need download artifacts from an external project set `GITLAB_PRIVATE_TOKEN` in GitLab variables and project and job to download.
Download all artifacts for job:
```yaml
...
simple test:
  stage: download
  image: gitlab-tool:latest
  script:
    - gitlab-tool download
        --project-id $EXT_PROJECT
        --destination local_path/
        --job-name $EXT_JOB_NAME
        --create-dirs
    - ls -l local_path/
...
```
For single file:
```yaml
...
simple test:
  stage: download
  image: gitlab-tool:latest
  script:
    - gitlab-tool download
        --project-id $EXT_PROJECT
        --destination local_path/
        --job-name $EXT_JOB_NAME
        --create-dirs
        --file-name $EXT_FILE_NAME
    - cat local_path/${EXT_FILE_NAME}
...
```