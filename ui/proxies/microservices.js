import axios from '~/plugins/back-office'
import { create as createClient } from '~/proxies/back-office-client'

class Microservices {
  constructor (client) {
    this.client = client
  }

  async one (id) {
    return await this.client.getOne(`/api/v1/microservices/${id}`)
  }

  async many (parameters = {}) {
    return await this.client.getMany('/api/v1/microservices', parameters)
  }
}

const createMicroservices = () => {
  return new Microservices(createClient(axios))
}

export default createMicroservices()
