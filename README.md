Hue Remote API Debugger
-------

# Setup
First setup an application at developers.meethue.com, then configure the clientID/Secret:

```bash
cp app-with-secrets.example.yaml app-with-secrets.yaml
vim app-with-secrets.yaml
```

# Deploying
```bash
gcloud app deploy app-with-secrets.yaml --project PROJECT_ID
```

# Credits

- https://cloud.google.com/appengine/docs/flexible/go/quickstart#deploy_and_run_hello_world_on_app_engine
- https://developers.meethue.com/documentation/getting-started
- Great help: https://medium.com/@pliutau/getting-started-with-oauth2-in-go-2c9fae55d187
