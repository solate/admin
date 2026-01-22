import type { Plugin } from 'vue'
import clickOutside from './clickOutside'

const directives: Plugin = {
  install(app) {
    app.directive('click-outside', clickOutside)
  }
}

export default directives
