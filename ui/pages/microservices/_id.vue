<template>
  <div>
    <v-card v-if="microservice" style="margin-bottom: 50px;">
      <v-card-title class="subheading font-weight-bold">
        Microservice: {{ microservice.attributes.name }}
      </v-card-title>
      <v-card-subtitle>
        <strong>{{ microservice.id }}</strong>
      </v-card-subtitle>

      <v-divider></v-divider>

      <v-list dense>
        <v-list-item>
          <v-list-item-content>&emsp;Description:</v-list-item-content>
          <v-list-item-content class="align-end">{{ microservice.attributes.description }}</v-list-item-content>
        </v-list-item>
        <v-list-item>
          <v-list-item-content>&emsp;Created at</v-list-item-content>
          <v-list-item-content class="align-end">{{ microservice.attributes.createdAt }}</v-list-item-content>
        </v-list-item>
      </v-list>
    </v-card>

    <h2>Service data entities</h2>

    <v-row dense>
      <v-col v-for="entity in entities" cols="4" :key="entity.id">
        <v-card
          color="#385F73"
          dark
        >
          <v-card-title class="headline">{{ entity.attributes.name }}</v-card-title>

          <v-card-subtitle>{{ entity.attributes.name }}</v-card-subtitle>

          <v-card-actions>
            <v-btn @click="updatePropertiesFor(entity)" text>Properties</v-btn>
          </v-card-actions>

          <v-card-text v-if="entity.attributes.properties">
            <v-list>
              <v-list-item v-for="property in entity.attributes.properties" :key="property.id" dense>
                <v-list-item-content>Name: {{ property.attributes.name }}</v-list-item-content>
                <v-list-item-content>
                  Changes:
                  <nuxt-link :to="{ name: 'events', query: { propertyId: property.id }}">
                    <v-chip
                      link="true"
                      class="ma-2"
                      color="secondary"
                    >
                      {{ property.attributes.changeCount }}
                    </v-chip>
                  </nuxt-link>
                </v-list-item-content>
              </v-list-item>
            </v-list>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </div>
</template>

<script>
  import microservices from "~/proxies/microservices";
  import entities from "~/proxies/entities";
  import properties from "~/proxies/properties";

  export default {
    name: "MicroservicePage",

    async fetch() {
      try {
        this.microservice = await microservices.one(this.$route.params.id);
        await this.fetchEntities(this.$route.params.id)
      } catch (e) {
        console.log(e)
      }
    },

    data: () => ({
      microservice: null,
      entities: [],
    }),

    methods: {
      async fetchEntities(serviceId) {
        const { collection, meta } = await entities.many({filter:{serviceId: serviceId}});
        console.log(collection, meta);
        this.entities = collection;
      },

      async updatePropertiesFor(entity) {
        try {
          const { collection, meta } = await properties.many({filter:{entityId: entity.id}});

          if (collection.length > 0) {
            this.entities = this.entities.map(e => {
              if (e.id === entity.id) {
                  e.attributes.properties = collection
              }

              return e
            })

            console.log(this.entities)
          }
        } catch (e) {
          console.log(e);
        }
      }
    }
  }
</script>

<style scoped>

</style>
