import axios from '~/plugins/back-office'
import { create as createClient } from '~/proxies/back-office-client'

class Properties {
  constructor (client) {
    this.client = client
  }

  async one (id) {
    return await this.client.getOne(`/api/v1/properties/${id}`)
  }

  async many (parameters = {}) {
    return await this.client.getMany('/api/v1/properties', parameters)
  }
}

const createEvents = () => {
  return new Properties(createClient(axios))
}

export default createEvents()
