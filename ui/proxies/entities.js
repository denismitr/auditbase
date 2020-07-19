import axios from '~/plugins/back-office'
import { create as createClient } from '~/proxies/back-office-client'

class Entities {
  constructor (client) {
    this.client = client
  }

  async one (id) {
    return await this.client.getOne(`/api/v1/entities/${id}`)
  }

  async many (parameters = {}) {
    return await this.client.getMany('/api/v1/entities', parameters)
  }
}

const createEvents = () => {
  return new Entities(createClient(axios))
}

export default createEvents()
