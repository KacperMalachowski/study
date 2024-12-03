#include <string>
#include <random>
#include <Arduino.h>
#include <PicoMQTT.h>
#include <DHT.h>
#include <ArduinoJson.h>
#include "secrets.h"

#define DHTPIN 0
#define MQTT_BROKER "mqtt.eclipseprojects.io"

using namespace std;


PicoMQTT::Client mqtt("mqtt.eclipseprojects.io");
DHT dht;
String uuid = String(random(0xffff), HEX);
int readingPeriod;

String prepareDiscoveryMsg() {
  JsonDocument res;
  res["idx"] = uuid;
  res["mqtt"] = "tir_us/" + uuid;

  JsonObject tempSensor = res["sensors"].add<JsonObject>();
  tempSensor["name"] = "Temperature";
  tempSensor["mqtt"] = "temperature";

  JsonObject humSensor = res["sensors"].add<JsonObject>();
  humSensor["name"] = "Humidity";
  humSensor["mqtt"] = "humidity";

  JsonObject setReadingPeriod = res["commands"].add<JsonObject>();
  setReadingPeriod["name"] = "setReadingPeriod";
  setReadingPeriod["type"] = "integer";
  setReadingPeriod["min"] = dht.getMinimumSamplingPeriod();

  String out;
  res.shrinkToFit();

  serializeJson(res, out);
  return out;
}

void setup() {
  // Setup sensors
  dht.setup(DHTPIN, DHT::DHT11);
  readingPeriod = dht.getMinimumSamplingPeriod();

  // Setup serial
  Serial.begin(115200);

  // Setup WiFi
  WiFi.begin(SECRET_SSID, SECRET_PASS);
  while (WiFi.status() != WL_CONNECTED) {
    delay(500);
    Serial.println("Connecting to WiFi...");

    if (WiFi.status() == WL_CONNECT_FAILED) {
      WiFi.begin(SECRET_FALLBACK_SSID, SECRET_FALLBACK_PASS);
    }
  }
  Serial.println("Connected to WiFi: " + WiFi.SSID() + " with IP: " + WiFi.localIP().toString());

  // Setup MQTT
  mqtt.subscribe("tir_us/discovery", [](const char * topic, const char * payload) {
    // Check if the payload contains discovery command in format {"command": "discover"}
    JsonDocument req;
    DeserializationError error = deserializeJson(req, payload);
    if (error) {
      String err = String(error.c_str());
      Serial.println("Failed to parse discovery request" + err);
      mqtt.publish("tir_us/" + uuid, "{\"error\": \"Failed to parse discovery request "+ err +"\"}");
      return;
    }
  
    if (req["command"] == "discover") {
      String out = prepareDiscoveryMsg();
      
      mqtt.publish("tir_us/discovery", out);
    }
  });

  mqtt.subscribe("tir_us/" + uuid, [](const char * topic, const char * payload) {
    JsonDocument req;
    DeserializationError error = deserializeJson(req, payload);
    if (error) {
      String err = String(error.c_str());
      Serial.println("Failed to parse command request" + err);
      mqtt.publish(topic, "{\"error\": \"Failed to parse command request "+ err +"\"}");
      return;
    }

    if (req["command"] == "setReadingPeriod") {
      int newReadingPeriod = req["value"];
      if (newReadingPeriod < dht.getMinimumSamplingPeriod()) {
        mqtt.publish(topic, "{\"error\": \"Reading period cannot be less than " + String(dht.getMinimumSamplingPeriod()) + "\"}");
        return;
      }

      readingPeriod = newReadingPeriod;
      mqtt.publish(topic, "{\"status\": \"Reading period set to " + String(readingPeriod) + "\", \"value\": " + String(readingPeriod) + "}");
    }

    if (req["command"] == "discover") {
      String out = prepareDiscoveryMsg();
      
      mqtt.publish(topic, out);
    }
  });

  mqtt.begin();
}

void loop() {
  mqtt.loop();

  float temperature = dht.getTemperature();
  float humidity = dht.getHumidity();

  DHT::DHT_ERROR_t status = dht.getStatus();
  if (status != DHT::ERROR_NONE) {
    Serial.println("Failed to read from DHT sensor, status: " + status);
    delay(dht.getMinimumSamplingPeriod());
    return;
  }


  Serial.println("Temperature: " + String(temperature) + "°C, Humidity: " + String(humidity) + "%");

  JsonDocument tempDoc;
  tempDoc["temperature"] = temperature;
  tempDoc["unit"] = "C";

  JsonDocument humDoc;
  humDoc["humidity"] = humidity;
  humDoc["unit"] = "%";

  char tempBuffer[1024];
  char humBuffer[1024];

  serializeJson(tempDoc, tempBuffer);
  serializeJson(humDoc, humBuffer);

  mqtt.publish("tir_us/" + uuid + "/temperature", tempBuffer);
  mqtt.publish("tir_us/" + uuid + "/humidity", humBuffer);

  delay(readingPeriod);
}
