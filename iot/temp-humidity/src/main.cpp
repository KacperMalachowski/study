#include <string>
#include <random>
#include <Arduino.h>
#include <PubSubClient.h>
#include <ESP8266WiFi.h>
#include <DHT.h>
#include "secrets.h"

#define DHTPIN 0
#define MQTT_BROKER "167.235.62.0"
#define DEVICE_ID "Temp_Humidity_04474"
#define MQTT_PREF "tir_us/" + string(DEVICE_ID)

using namespace std;

WiFiClient espClient;
PubSubClient client(espClient);
DHT dht;

void setup_wifi() {
  WiFi.begin(SECRET_SSID, SECRET_PASS);
  while (WiFi.status() != WL_CONNECTED) {
    delay(500);
    Serial.print(".");
  }
  Serial.println("WiFi connected");
  Serial.printf("IP address: %s\n", WiFi.localIP().toString().c_str());
}

void reconnect() {
  while (!client.connected()) {
    Serial.print("Attempting MQTT connection...");
    if (client.connect("Temp_Humidity", MQTT_USER, MQTT_PASS)) {
      Serial.println("connected");
    } else {
      Serial.printf("failed, rc=%d try again in 5 seconds\n", client.state());
      delay(5000);
    }
  }
}

void setup() {
  Serial.begin(9600);
  setup_wifi();
  client.setServer(MQTT_BROKER, 1883);

  dht.setup(DHTPIN, DHT::DHT11);
}

void loop() {
  char msg[8];
  if (!client.connected()) {
    reconnect();
  }

  float humidity = dht.getHumidity();
  float temperature = dht.getTemperature();

  snprintf(msg, 8, "%f", humidity);
  client.publish((MQTT_PREF + "/humidity").c_str(), msg);
  snprintf(msg, 8, "%f", temperature);
  client.publish((MQTT_PREF + "/temperature").c_str(), msg);

  Serial.printf("Humidity: %f\n", humidity);
  Serial.printf("Temperature: %f\n", temperature);

  delay(dht.getMinimumSamplingPeriod());
}

