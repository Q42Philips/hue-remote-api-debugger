Hue Remote API Debugger
-------

# Local setup
First setup an application at developers.meethue.com, then configure the clientID/Secret:

```bash
go build
CALLBACK_URL="http://localhost:8080/hue_callback_url" \
HUE_CLIENT_ID=myclientid \
HUE_CLIENT_SECRET=myclientsecret \
HUE_APPID="myappid" \
./hue-remote-api-debugger
```

# App Engine flex setup
First setup an application at developers.meethue.com, then configure the clientID/Secret:

```bash
cp app-with-secrets.example.yaml app-with-secrets.yaml
vim app-with-secrets.yaml
gcloud app deploy app-with-secrets.yaml --project PROJECT_ID
```

# Credits

- https://cloud.google.com/appengine/docs/flexible/go/quickstart#deploy_and_run_hello_world_on_app_engine
- https://developers.meethue.com/documentation/getting-started
- Great help: https://medium.com/@pliutau/getting-started-with-oauth2-in-go-2c9fae55d187
