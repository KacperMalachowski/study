#include <string>
#include <random>
#include <Arduino.h>
#include <PicoMQTT.h>
#include <DHT.h>
#include <ArduinoJson.h>
#include "UUID.h"
#include "secrets.h"

#define DHTPIN 0
#define DHTTYPE DHT11
#define MQTT_BROKER "mqtt.eclipseprojects.io"

using namespace std;

PicoMQTT::Client mqtt("mqtt.eclipseprojects.io");
DHT dht(DHTPIN, DHTTYPE);
String uuid = UUID().toCharArray();

void setup() {
  Serial.begin(115200);

  // Connect to WiFi, fallback to fallback SSID and password on failure
  WiFi.begin(SECRET_SSID, SECRET_PASS);
  while (WiFi.status() != WL_CONNECTED) {
    delay(500);
    Serial.println("Connecting to WiFi...");

    if (WiFi.status() == WL_CONNECT_FAILED) {
      WiFi.begin(SECRET_FALLBACK_SSID, SECRET_FALLBACK_PASS);
    }
  }
  Serial.println("Connected to WiFi: " + WiFi.SSID() + " with IP: " + WiFi.localIP().toString());

  // Subscribe to a discovery topic
  mqtt.subscribe("tir_us/discovery", [](const char * topic, const char * payload) {
    JsonDocument doc;
    DeserializationError error = deserializeJson(doc, payload);
    if (error) {
      Serial.println("Failed to parse JSON " + String(error.c_str()));
      return;
    }

    if (doc["command"] == "discover") {
      JsonDocument response;
      JsonObject device = response["device"].to<JsonObject>();
      device["name"] = "ESP8266";
      device["manufacturer"] = "Espressif";
      device["idx"] = WiFi.macAddress().c_str();
      device["mqttTopicPrefix"] = "tir_us/" + uuid;
      
      JsonObject sensors = response["sensors"].to<JsonObject>();
      JsonObject temperature = sensors["temperature"].to<JsonObject>();
      temperature["name"] = "Temperature";
      temperature["mqttTopic"] = "temperature";

      JsonObject humidity = sensors["humidity"].to<JsonObject>();
      humidity["name"] = "Humidity";
      humidity["mqttTopic"] = "humidity";

      response["commands"].to<JsonArray>();

      char buffer[1024];
      serializeJson(response, buffer);

      mqtt.publish("tir_us/discovery", buffer);
    }
  });

  mqtt.begin();
}

void loop() {
  mqtt.loop();

  float temperature = dht.readTemperature();
  float humidity = dht.readHumidity();

  if (isnan(temperature) || isnan(humidity)) {
    Serial.println("Failed to read from DHT sensor!");
    delay(1000);
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

  delay(5000);
}
