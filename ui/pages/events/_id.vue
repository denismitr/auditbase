<template>
  <div>
    <event-card v-if="event" :event="event"/>
    <span v-else>Foo</span>
  </div>
</template>

<script>
  import events from './../../proxies/events';
  import EventCard from "~/components/EventCard";

  export default {
    name: 'EventId',

    components: { EventCard },

    async fetch() {
      console.log(this.$route.params.id)
      try {
        const data = await events.one(this.$route.params.id)
        this.event = data;
      } catch (e) {
        console.log(e)
      }
    },

    data () {
      return {
        event: null,
      };
    }
  }
</script>

<style scoped>

</style>
