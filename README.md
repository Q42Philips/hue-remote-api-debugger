Hue Remote API Debugger
-------

This application can be used to send [CLIP commands](https://developers.meethue.com/documentation/getting-started) to your Philips Hue bridge via its cloud connection to manipulate the lights in your home.

# Local setup
- First setup an application at [developers.meethue.com](https://developers.meethue.com/user/me/apps)
- Then configure the proclaimed ClientID/Secret when running the app:

```bash
CALLBACK_URL="http://localhost:8080/hue_callback_url" \
HUE_CLIENT_ID=myclientid \
HUE_CLIENT_SECRET=myclientsecret \
HUE_APPID="myappid" \
go run ./cmd
```

# App Engine flex setup
- First setup an application at [developers.meethue.com](https://developers.meethue.com/user/me/apps)
- Then configure the proclaimed ClientID/Secret in `app-with-secrets.yaml` and tweak the defaults:

```bash
cp app-with-secrets.example.yaml app-with-secrets.yaml
vim app-with-secrets.yaml
```

- finally deploy to google: `gcloud app deploy app-with-secrets.yaml --project PROJECT_ID`

# Credits

- https://cloud.google.com/appengine/docs/flexible/go/quickstart#deploy_and_run_hello_world_on_app_engine
- https://developers.meethue.com/documentation/getting-started
- Great help: https://medium.com/@pliutau/getting-started-with-oauth2-in-go-2c9fae55d187
