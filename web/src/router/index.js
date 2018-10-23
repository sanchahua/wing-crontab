import Vue from 'vue'
import Router from 'vue-router'
import HelloWorld from '@/components/HelloWorld'
import Index from '@/components/pages/Index'
import Add from '@/components/pages/Add'
import CronList from '@/components/pages/CronList'
import Edit from '@/components/pages/Edit'
import Logs from '@/components/pages/Logs'
import LogDetail from '@/components/pages/LogDetail'
import UserAdd from '@/components/pages/UserAdd'
import Users from '@/components/pages/Users'
import UserEdit from '@/components/pages/UserEdit'
import UserCenter from '@/components/pages/UserCenter'
import UserPowers from '@/components/pages/UserPowers'
import Nodes from '@/components/pages/Nodes'

Vue.use(Router)

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
    },
    {
      path: '/edit',
      name: 'Edit',
      component: Edit,
    },
    {
      path: '/logs',
      name: 'Logs',
      component: Logs,
    }
    ,
    {
      path: '/log_detail',
      name: 'LogDetail',
      component: LogDetail,
    },
    {
      path: '/user_add',
      name: 'UserAdd',
      component: UserAdd,
    },
    //Users
    {
      path: '/users',
      name: 'Users',
      component: Users,
    },
    {
      path: '/user_edit',
      name: 'UserEdit',
      component: UserEdit,
    },
    {
      path: '/user_center',
      name: 'UserCenter',
      component: UserCenter,
    },
    {
      path: '/user_powers',
      name: 'UserPowers',
      component: UserPowers,
    },
    {
      path: 'nodes',
      name: 'Nodes',
      component: Nodes,
    }
  ]
})
