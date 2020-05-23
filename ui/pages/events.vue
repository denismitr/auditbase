<template>
  <v-layout>
    <v-flex class="text-center">
      <img
        src="/v.png"
        alt="Vuetify.js"
        class="mb-5"
      >
      <blockquote class="blockquote">
        &#8220;Let's load events.&#8221;
        <footer>
          <small>
            <em>&mdash;DM</em>
          </small>
        </footer>
      </blockquote>
    </v-flex>
  </v-layout>
</template>

<script>
import events from '~/proxies/events'

export default {
  async fetch () {
    const { collection, meta } = await events.many()
    this.events = collection
    this.meta = meta
  },

  data () {
    return {
      events: [],
      meta: {}
    }
  }
}
</script>
