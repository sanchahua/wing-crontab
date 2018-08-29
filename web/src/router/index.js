import Vue from 'vue'
import Router from 'vue-router'
import HelloWorld from '@/components/HelloWorld'
import Index from '@/components/pages/Index'
import Add from '@/components/pages/Add'
import CronList from '@/components/pages/CronList'

Vue.use(Router)
//
// const routes = {
//   '/': Home,
//   '/about': About
// }
//
// new Vue({
//   el: '#app',
//   data: {
//     currentRoute: window.location.pathname
//   },
//   computed: {
//     ViewComponent () {
//       return routes[this.currentRoute] || NotFound
//     }
//   },
//   render (h) { return h(this.ViewComponent) }
// })

export default new Router({
  routes: [
    {
      path: '/',
      name: 'Index',
      component: Index
    },
    {
      path: '/add',
      name: "Add",
      component: Add
    },
    {
      path: '/cron_list',
      name: 'CronList',
      component: CronList,
    }
  ]
})
