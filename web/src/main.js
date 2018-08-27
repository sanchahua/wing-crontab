// The Vue build version to load with the `import` command
// (runtime-only or standalone) has been set in webpack.base.conf with an alias.
import Vue from 'vue'
import App from './App'
import router from './router'

//import $ from 'jquery'

import Index from '@/components/pages/Index'
import Forms from '@/components/pages/Forms'

const routes = {
  '#/': Index,
  '#/froms': Forms
}

Vue.config.productionTip = false
const NotFound = { template: '<p>Page not found</p>' }
/* eslint-disable no-new */
new Vue({
  el: '#outter-wp',
  // data: {
  //   currentRoute: window.location.hash
  // },
  // computed: {
  //   ViewComponent () {
  //     console.log(this.currentRoute);
  //     return routes[this.currentRoute] || NotFound
  //   }
  // },
  // render (h) { return h(this.ViewComponent) }
  router,
  components: { App },
  template: '<App/>'
})
