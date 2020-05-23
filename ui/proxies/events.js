import axios from '~/plugins/back-office'
import { create as createClient } from '~/proxies/back-office-client'

class Events {
  constructor (client) {
    this.client = client
  }

  async one (id) {
    return await this.client.getOne(`/api/v1/events/${id}`)
  }

  async many (parameters = {}) {
    return await this.client.getMany('/api/v1/events', parameters)
  }
}

const createEvents = () => {
  return new Events(createClient(axios))
}

export default createEvents()
