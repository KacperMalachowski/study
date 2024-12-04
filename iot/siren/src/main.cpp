#include <string>
#include <Arduino.h>
#include <PubSubClient.h>
#include <ESP8266WiFi.h>
#include "secrets.h"

#define LED_PIN 0
#define DEVICE_ID "Alarm_ecbf8"
#define MQTT_BROKER "167.235.62.0"
#define MQTT_PREF "tir_us/" + string(DEVICE_ID)

using namespace std;

WiFiClient espClient;
PubSubClient client(espClient);
int ledState = LOW;

void setup_wifi()
{
  WiFi.begin(SECRET_SSID, SECRET_PASS);
  while (WiFi.status() != WL_CONNECTED)
  {
    delay(500);
    Serial.print(".");
  }
  Serial.println("WiFi connected");
  Serial.printf("IP address: %s\n", WiFi.localIP().toString().c_str());
}

void reconnect(){
  while (!client.connected()) {
    Serial.print("Attempting MQTT connection...");
    if (client.connect("Alarm", MQTT_USER, MQTT_PASS)) {
      Serial.println("connected");
      client.subscribe((MQTT_PREF + "/led").c_str());
    } else {
      Serial.printf("failed, rc=%d try again in 5 seconds\n", client.state());
      delay(5000);
    }
  }
}

void callback(char *topic, uint8_t *payload, unsigned int length)
{
  Serial.print("Message arrived [");
  Serial.print(topic);
  Serial.print("] ");
  for (int i = 0; i < length; i++)
  {
    Serial.print((char)payload[i]);
  }
  Serial.println();

  if (string(topic) == (MQTT_PREF + "/led"))
  {
    if ((char)payload[0] == '1')
    {
      ledState = HIGH;
      Serial.println("LED ON");
    }
    else
    {
      ledState = LOW;
      Serial.println("LED OFF");
    }
    digitalWrite(LED_PIN, ledState);
  }
}

void setup() {
  pinMode(LED_PIN, OUTPUT);
  Serial.begin(9600);
  setup_wifi();
  client.setServer(MQTT_BROKER, 1883);
  client.setCallback(callback);

}

void loop() {
  char msg[8];
  if (!client.connected())
  {
    reconnect();
  }
  client.loop();
  delay(100);
}