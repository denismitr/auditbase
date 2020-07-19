<template>
  <div>
    <v-card style="margin-bottom: 50px;">
      <v-card-title class="subheading font-weight-bold">
        Event: {{ event.attributes.eventName }}
      </v-card-title>
      <v-card-subtitle>
        <strong>{{ event.id }}</strong>
      </v-card-subtitle>

      <v-divider></v-divider>

      <v-card-title class="subheading font-weight-bold">
        Actor
      </v-card-title>

      <v-list dense>
        <v-list-item>
          <v-list-item-content>&emsp;ID</v-list-item-content>
          <v-list-item-content class="align-end">{{ event.attributes.actorId }}</v-list-item-content>
        </v-list-item>
        <v-list-item>
          <v-list-item-content>Actor Entity:</v-list-item-content>
        </v-list-item>
        <v-list-item>
          <v-list-item-content>&emsp;ID</v-list-item-content>
          <v-list-item-content class="align-end">{{ event.attributes.actorEntity.id }}</v-list-item-content>
        </v-list-item>
        <v-list-item>
          <v-list-item-content>&emsp;Name</v-list-item-content>
          <v-list-item-content class="align-end">{{ event.attributes.actorEntity.name }}</v-list-item-content>
        </v-list-item>
        <v-list-item v-if="event.attributes.actorEntity.description">
          <v-list-item-content>&emsp;Description</v-list-item-content>
          <v-list-item-content class="align-end">{{ event.attributes.actorEntity.description }}</v-list-item-content>
        </v-list-item>
        <v-list-item>
          <v-list-item-content>Actor Service:</v-list-item-content>
        </v-list-item>
        <v-list-item>
          <v-list-item-content>&emsp;ID</v-list-item-content>
          <v-list-item-content class="align-end">
            <nuxt-link :to="{name: 'microservices-id', params: {id: event.attributes.actorService.id}}">
              <v-chip
                      link="true"
                      class="ma-2"
                      color="primary"
              >
                {{ event.attributes.actorService.id }}
              </v-chip>
            </nuxt-link>
          </v-list-item-content>
        </v-list-item>
        <v-list-item>
          <v-list-item-content>&emsp;Name</v-list-item-content>
          <v-list-item-content class="align-end">{{ event.attributes.actorService.name }}</v-list-item-content>
        </v-list-item>
      </v-list>

      <v-card-title class="subheading font-weight-bold">
        Target
      </v-card-title>

      <v-list dense>
        <v-list-item>
          <v-list-item-content>&emsp;ID</v-list-item-content>
          <v-list-item-content class="align-end">{{ event.attributes.targetId }}</v-list-item-content>
        </v-list-item>
        <v-list-item>
          <v-list-item-content>Target Entity:</v-list-item-content>
        </v-list-item>
        <v-list-item>
          <v-list-item-content>&emsp;ID</v-list-item-content>
          <v-list-item-content class="align-end">{{ event.attributes.targetEntity.id }}</v-list-item-content>
        </v-list-item>
        <v-list-item>
          <v-list-item-content>&emsp;Name</v-list-item-content>
          <v-list-item-content class="align-end">{{ event.attributes.targetEntity.name }}</v-list-item-content>
        </v-list-item>
        <v-list-item>
          <v-list-item-content>Target Service:</v-list-item-content>
        </v-list-item>
        <v-list-item>
          <v-list-item-content>&emsp;ID</v-list-item-content>
          <v-list-item-content class="align-end">
            <nuxt-link :to="{name: 'microservices-id', params: {id: event.attributes.targetService.id}}">
              <v-chip
                      link="true"
                      class="ma-2"
                      color="primary"
              >
                {{ event.attributes.targetService.id }}
              </v-chip>
            </nuxt-link>
          </v-list-item-content>
        </v-list-item>
        <v-list-item>
          <v-list-item-content>&emsp;Name</v-list-item-content>
          <v-list-item-content class="align-end">{{ event.attributes.targetService.name }}</v-list-item-content>
        </v-list-item>
      </v-list>

      <v-card-title class="subheading font-weight-bold">
        Timing
      </v-card-title>

      <v-list dense>
        <v-list-item>
          <v-list-item-content>&emsp;Emitted At</v-list-item-content>
          <v-list-item-content class="align-end">{{ event.attributes.emittedAt }}</v-list-item-content>
        </v-list-item>
        <v-list-item>
          <v-list-item-content>&emsp;Registered at auditbase</v-list-item-content>
          <v-list-item-content class="align-end">{{ event.attributes.registeredAt }}</v-list-item-content>
        </v-list-item>
      </v-list>
    </v-card>

    <v-card class="">
      <v-card-title class="subheading font-weight-bold">
        Changes
      </v-card-title>

      <div style="padding: 20px;" v-for="(change) in event.attributes.changes">
        <v-divider></v-divider>
        <v-list dense>
          <v-list-item>
            <v-list-item-content>&emsp;Property</v-list-item-content>
            <v-list-item-content class="align-end">{{ change.property }}</v-list-item-content>
          </v-list-item>
          <v-list-item>
            <v-list-item-content>&emsp;From</v-list-item-content>
            <v-list-item-content class="align-end">{{ change.from }}</v-list-item-content>
          </v-list-item>
          <v-list-item>
            <v-list-item-content>&emsp;To</v-list-item-content>
            <v-list-item-content class="align-end">{{ change.to }}</v-list-item-content>
          </v-list-item>
        </v-list>
      </div>
    </v-card>
  </div>
</template>

<script>
    export default {
        name: "event",

        props: {
          event: {
            type: Object,
            required: true,
          }
        }
    }
</script>

<style scoped>

</style>
