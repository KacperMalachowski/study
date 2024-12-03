#include <string>
#include <random>
#include <Arduino.h>
#include <PicoMQTT.h>
#include <ArduinoJson.h>
#include "secrets.h"

#define SIREN_PIN 0
#define MQTT_BROKER "mqtt.eclipseprojects.io"

using namespace std;

PicoMQTT::Client mqtt("mqtt.eclipseprojects.io");
String uuid = String(random(0xffff), HEX);

String prepareDiscoveryMsg() {
  JsonDocument res;
  res["idx"] = uuid;
  res["mqtt"] = "tir_us/" + uuid;

  JsonObject siren = res["commands"].createNestedObject();
  siren["name"] = "siren";
  siren["type"] = "boolean";

  JsonObject sirenStatus = res["getters"].createNestedObject();
  sirenStatus["name"] = "sirenStatus";

  String out;
  res.shrinkToFit();

  serializeJson(res, out);
  return out;
}

void setup() {
  // Setup sensors
  pinMode(SIREN_PIN, OUTPUT);

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

    if (req["command"] == "discover") {
      String out = prepareDiscoveryMsg();
      
      mqtt.publish(topic, out);
    }

    if (req["command"] == "siren") {
      bool value = req["value"];
      digitalWrite(SIREN_PIN, value == 1 ? HIGH : LOW);
      mqtt.publish(topic, "{\"status\": \"ok\", \"value\": " + String(value) + "}");
    }

    if(req["command"] == "get") {
      if(req["name"] == "sirenStatus") {
        bool status = digitalRead(SIREN_PIN);
        mqtt.publish(topic, "{\"status\": \"ok\", \"value\": " + String(status) + "}");
      }
    }
  });

  mqtt.begin();
}

void loop() {
  mqtt.loop();
}