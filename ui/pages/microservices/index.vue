<template>
  <v-row class="pa-6">
    <v-col cols="12">
            <v-data-iterator
          :items="collection"
          :items-per-page.sync="cursor.perPage"
          :page="cursor.page"
          :search="cursor.search"
          :sort-by="cursor.sortBy.toLowerCase()"
          :sort-desc="cursor.sortDesc"
          hide-default-footer
        >
          <template v-slot:header>
            <v-toolbar
              dark
              color="blue darken-3"
              class="mb-1"
            >
              <v-text-field
                v-model="cursor.search"
                clearable
                flat
                solo-inverted
                hide-details
                prepend-inner-icon="mdi-magnify"
                label="Search microservices by"
              ></v-text-field>
              <template v-if="$vuetify.breakpoint.mdAndUp">
                <v-spacer></v-spacer>
                <v-select
                  v-model="cursor.sortBy"
                  flat
                  solo-inverted
                  hide-details
                  :items="['name', 'id']"
                  label="Sort by"
                ></v-select>
                <v-spacer></v-spacer>
                <v-btn-toggle
                  v-model="cursor.sortDesc"
                  mandatory
                >
                  <v-btn
                    large
                    depressed
                    color="blue"
                    :value="false"
                  >
                    <v-icon>mdi-arrow-up</v-icon>
                  </v-btn>
                  <v-btn
                    large
                    depressed
                    color="blue"
                    :value="true"
                  >
                    <v-icon>mdi-arrow-down</v-icon>
                  </v-btn>
                </v-btn-toggle>
              </template>
            </v-toolbar>
          </template>

          <template v-slot:default="props">
            <v-row>
              <v-col
                v-for="item in props.items"
                :key="item.name"
                cols="12"
                sm="12"
                md="6"
                lg="6"
              >
                <v-card>
                  <v-card-title class="subheading font-weight-bold">
                    {{ item.attributes.name }}
                  </v-card-title>
                  <v-card-subtitle>
                    <nuxt-link :to="'/microservices/' + item.id">{{ item.id }}</nuxt-link>
                  </v-card-subtitle>

                  <v-divider></v-divider>

                  <v-list dense>
                    <v-list-item
                      v-for="(key, index) in columns"
                      :key="index"
                    >
                      <v-list-item-content :class="{ 'blue--text': cursor.sortBy === key }">{{ key }}:</v-list-item-content>
                      <v-list-item-content class="align-end" :class="{ 'blue--text': cursor.sortBy === key }">{{ item.attributes[key] }}</v-list-item-content>
                    </v-list-item>
                  </v-list>
                </v-card>
              </v-col>
            </v-row>
          </template>

          <template v-slot:footer>
            <v-row class="mt-2" align="center" justify="center">
              <span class="grey--text">Items per page</span>
              <v-menu offset-y>
                <template v-slot:activator="{ on }">
                  <v-btn
                    dark
                    text
                    color="primary"
                    class="ml-2"
                    v-on="on"
                  >
                    {{ cursor.perPage }}
                    <v-icon>mdi-chevron-down</v-icon>
                  </v-btn>
                </template>
                <v-list>
                  <v-list-item
                    v-for="(number, index) in cursor.perPageArray"
                    :key="index"
                    @click="updatePerPage(number)"
                  >
                    <v-list-item-title>{{ number }}</v-list-item-title>
                  </v-list-item>
                </v-list>
              </v-menu>

              <v-spacer></v-spacer>

              <span
                class="mr-4
              grey--text"
              >
              Page {{ cursor.page }} of {{ numberOfPages }}
            </span>
              <v-btn
                fab
                dark
                color="blue darken-3"
                class="mr-1"
                @click="prevPage"
              >
                <v-icon>mdi-chevron-left</v-icon>
              </v-btn>
              <v-btn
                fab
                dark
                color="blue darken-3"
                class="ml-1"
                @click="nextPage"
              >
                <v-icon>mdi-chevron-right</v-icon>
              </v-btn>
            </v-row>
          </template>
        </v-data-iterator>
          </v-col>
  </v-row>
</template>

<script>
  import microservices from "~/proxies/microservices";

  export default {
    name: "microservice.index",

    computed: {
      numberOfPages () {
        return Math.ceil(this.collection.length / this.cursor.perPage)
      },
    },

    data () {
      return {
        collection: [],
        meta: {},
        filter: {},
        cursor: {
          sortDesc: false,
          page: 1,
          perPage: 10,
          perPageArray: [10, 20, 30],
          sortBy: 'name',
        },
        columns: [
          'name',
          'description',
          'createdAt',
          'updatedAt',
        ]
      }
    },

    async fetch() {
      try {
          const { collection, meta } = await microservices.many();
          this.collection = collection;
          this.meta = meta;
      } catch (e) {
        console.log("ERROR: ", e)
      }
    },

    methods: {
      nextPage () {
        if (this.cursor.page + 1 <= this.cursor.numberOfPages) {
          this.cursor.page += 1
        }
      },

      prevPage () {
        if (this.cursor.page - 1 >= 1) {
          this.cursor.page -= 1
        }
      },

      updatePerPage (number) {
        this.cursor.perPage = number
      }
    }
  }
</script>

<style scoped>

</style>
