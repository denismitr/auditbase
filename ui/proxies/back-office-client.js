import queryString from 'qs'

class BackOfficeClient {
  constructor (axios, headers = {}) {
    this.headers = headers
    this.axios = axios
  }

  async getOne (endpoint, parameters = {}) {
    const response = await this.axios.get(this.buildUri(endpoint, parameters))
    const data = response.data.data

    return {
      id: data.id,
      attributes: data.attributes
    }
  }

  async getMany (endpoint, parameters = {}) {
    const response = await this.axios.get(this.buildUri(endpoint, parameters))
    const data = response.data.data
    const meta = response.data.meta

    return {
      collection: data,
      meta
    }
  }

  buildUri (endpoint, parameters = {}) {
    const qs = queryString.stringify(parameters, {encode: false})

    if (qs.length === 0) {
      return endpoint
    }

    return `${endpoint}?${qs}`
  }
}

const create = (axios, headers) => {
  return new BackOfficeClient(axios, headers)
}

export { create, BackOfficeClient }
