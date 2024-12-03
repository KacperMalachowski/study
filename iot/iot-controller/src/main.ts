import './assets/main.css'

import { createApp } from 'vue'
import App from './App.vue'
import { createPahoMqttPlugin } from 'vue-paho-mqtt'

createApp(App)
  .use(
    createPahoMqttPlugin({
      PluginOptions: {
        autoConnect: false,
        showNotifications: false,
      },
      MqttOptions: {
        host: 'mqtt.eclipseprojects.io',
        port: 80,
        clientId: `tir-us-${Math.random().toString(16).substring(2, 8)}`,
        mainTopic: 'tir_us',
        path: '/mqtt',
      }
    })
  )
  .mount('#app')
