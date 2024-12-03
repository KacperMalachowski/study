<script setup lang="ts">
import type { RefSymbol } from '@vue/reactivity'
import { onMounted, reactive, ref, useTemplateRef } from 'vue'
import { $mqtt } from 'vue-paho-mqtt'

type Device = {
  mqtt: string
  commands?: {
    name: string
    type: string
    min?: number
  }[]
  sensors?: {
    name: string
    type: string
    mqtt: string
    value?: any
    units?: string
  }[]
  getters?: {
    name: string
    type: string
    value?: any
  }[]
}

const devices = reactive<{ [key: string]: Device }>({})
const input = ref<number>(0)

function handleCommand(device: Device, command: string) {
  // Get value from input
  const cmd = device.commands?.find((c) => c.name === command)
  if (cmd?.type === 'integer') {
    if (input.value) {
      $mqtt.publish(device.mqtt, JSON.stringify({ command, value: input.value }), 'B', false)
    }
  } else {
    const status = device.getters?.find((g) => g.name.startsWith(cmd!.name))
    const message = JSON.stringify({ command, value: status?.value ? 0 : 1 })
    console.log(message, status)
    $mqtt.publish(device.mqtt, message, 'B', false)
  }
}

onMounted(() => {
  $mqtt.subscribe('discovery', (message) => {
    const payload = JSON.parse(message)

    console.log(payload, devices.value)
    if (payload.command) {
      return
    }

    const device: Device = {
      mqtt: payload.mqtt,
      commands: payload.commands,
      sensors: payload.sensors,
      getters: payload.getters,
    }

    device.sensors?.forEach((sensor) => {
      $mqtt.subscribe(`${payload.idx}/${sensor.mqtt}`, (message) => {
        const value = JSON.parse(message)
        if (!value.error) {
          sensor.value = value[sensor.mqtt] ?? value.value
          sensor.units = value.unit
        }
        if (value.error) {
          console.error(value.error)
        }

        devices[payload.idx].sensors = device.sensors?.map((s) => {
          if (s.mqtt === sensor.mqtt) {
            return sensor
          }
          return s
        })
      })
    })

    device.getters?.forEach((getter) => {
      $mqtt.subscribe(`${payload.idx}`, (message) => {
        const value = JSON.parse(message)
        if (!value.error) {
          getter.value = value[getter.name] ?? value.value
        }
        if (value.error) {
          console.error(value.error)
        }

        devices[payload.idx].getters = device.getters?.map((g) => {
          if (g.name === getter.name) {
            return getter
          }
          return g
        })
      })

      $mqtt.publish(payload.idx, JSON.stringify({ command: 'get', name: getter.name }), 'B')
    })

    devices[payload.idx] = device
  })
  $mqtt.connect({
    onConnect: () => {
      $mqtt.publish('discovery', JSON.stringify({ command: 'discover' }), 'B')
    },
  })
})
</script>

<template>
  <div>
    <button @click="$mqtt.publish('discovery', JSON.stringify({ command: 'discover' }), 'B')">
      Refresh
    </button>
    <div v-for="device in devices" :key="device.mqtt">
      <h2>{{ device.mqtt }}</h2>
      <div v-for="sensor in device.sensors" :key="sensor.mqtt">
        <p>{{ sensor.name }}: {{ sensor.value }} {{ sensor.units }}</p>
      </div>
      <div v-for="command in device.commands" :key="command.name">
        <button @click="handleCommand(device, command.name)">
          {{ command.name }}
        </button>
        <input v-model="input" v-if="command.type === 'integer'" type="number" :min="command.min" />
      </div>
    </div>
  </div>
</template>

<style scoped>
button {
  padding: 0.5rem 1rem;
  border: 1px solid #000;
  border-radius: 0.25rem;
  background-color: #fff;
  color: #000;
  font-size: 1rem;
  cursor: pointer;

  &:hover {
    background-color: #000;
    color: #fff;
  }
}
</style>
