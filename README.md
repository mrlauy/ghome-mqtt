# gHome-MQTT

Cloud2Cloud server to send actions to your MQTT server.

## Setup

In the Google Actions Console configure the following:
- Fullfillment URL: https://exmaple.com/smarthome/fulfillment
- OAuth2 client id
- OAuth2 client secret
- Authorization URL: https://exmaple.com/oauth/authorize
- Token URL: https://exmaple.com/oauth/token

### Config
Create a `config.yaml` or override the location with the environment variable`CONFIG_FILE`. Config the client id and secret in the config file or environment variables.

### Devices
Create a `devices.json` with all the devices. This will be return when Google is trying to sync.

### Credentials
Create username and password credentials to login on the server: [How to Create credentials](credentials/README.md).

## Run

```shell
go run .
```

## References:
- [Google Actions Project](https://console.actions.google.com/project/smart-node-438/overview)
- [cloud-to-cloud Traits Documentation](https://developers.home.google.com/cloud-to-cloud/traits)
- [OAuth2 Requirements](https://developers.home.google.com/cloud-to-cloud/project/authorization)
- [Cloudflare Zero Trust Tunnels](https://one.dash.cloudflare.com/f4278dde21a39adcf34551b262ce286d/access/tunnels)

### Inspirations
- [HomeAutio.Mqtt.GoogleHome](https://github.com/i8beef/HomeAutio.Mqtt.GoogleHome)
- [smart-home-nodejs](https://github.com/google-home/smart-home-nodejs)
- [Zigbee2MQTT](https://www.zigbee2mqtt.io/guide/usage/mqtt_topics_and_messages.html#zigbee2mqtt-friendly-name-set)
 