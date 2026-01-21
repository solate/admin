import clickOutside from './clickOutside'

export default {
  install(app) {
    app.directive('click-outside', clickOutside)
  }
}
