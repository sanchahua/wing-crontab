import Vue from 'vue'
import Router from 'vue-router'
import HelloWorld from '@/components/HelloWorld'
import Index from '@/components/pages/Index'
import Forms from '@/components/pages/Forms'
Vue.use(Router)

export default new Router({
  routes: [
    {
      path: '/',
      name: 'Index',
      component: Index
    },
    {
      path: '/forms',
      name: "Forms",
      component: Forms
    }
  ]
})
