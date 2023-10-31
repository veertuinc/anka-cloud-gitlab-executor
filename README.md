# Gitlab Custom Executor

### How to run gitlab runner locally?
1. Fetch Registration token from your Gitlab repo -> Settings -> CI/CD -> Runners
2. Generate a config file:
    ./gitlab-runner register
3. Run the runner
    ./gitlab-runner run
4. Execute your pipeline from Gitlab