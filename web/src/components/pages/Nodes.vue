<template>
  <!--/sub-heard-part-->
  <div>
    <div class="sub-heard-part">
      <ol class="breadcrumb m-b-0">
        <li><a href="/ui/#/">首页</a></li>
        <li class="active">服务器集群</li>
      </ol>
    </div>
    <div class="tables">
      <table id="cron-list-table" class="table table-bordered" width="100%">
        <thead>
        <tr>
          <th class="sh-row">服务器</th>
          <th class="sh-row">服务地址</th>
          <th class="sh-row">状态</th>
          <th class="sh-row">Leader</th>
          <th class="sh-row">下线</th>
          <th>操作</th>
        </tr>
        </thead>
        <tbody>
        <tr v-for="item in nodes">
          <th class="sh-row">{{item.Name}}</th>
          <td class="sh-row">{{item.Address}}</td>

          <td class="sh-row" v-if="item.Status == 1">正常</td>
          <td class="sh-row" v-else>故障</td>

          <td class="sh-row" v-if="item.Leader">是</td>
          <td class="sh-row" v-else>否</td>
          <td class="sh-row" v-if="item.Offline">是</td>
          <td class="sh-row" v-else>否</td>
          <td>
            <div>
              <a class="btn" v-bind:item-id="item.ID" v-if="item.Offline" v-on:click="online">上线</a>
              <a class="btn" v-bind:item-id="item.ID" v-else v-on:click="offline">下线</a>
              <a class="btn" v-bind:item-id="item.ID" v-on:click="del">删除</a>
            </div>
          </td>
        </tr>
        </tbody>
      </table>
    </div>
  </div>
  <!--//forms-->
</template>
<script>
  export default {
    name: "Nodes",
    data: function() {
      return {
        nodes: [],
        datetime: (new Date()).Format("yyyy-MM-dd hh:mm:ss"),
      }
    },
    created: function() {
      this.getList()
    },
    methods: {
      getList: function() {
        let that = this
        axios.get('/services?time='+(new Date()).valueOf()).then(function (response) {
          if (2000 == response.data.code) {
            that.nodes = response.data.data
            console.log(response);
          } else if (8000 == response.data.code) {
            window.location.href="/ui/login.html"
          } else {
            alert(response.data.message);
          }
        }).catch(function (error) {

        });
      },
      online: function () {
        let id = $(event.target).attr("item-id");
        let that = this
        axios.post('/services/online/' + id + '?time='+(new Date()).valueOf()).then(function (response) {
          console.log(response);
          if (2000 == response.data.code) {
            for (let i = 0; i < that.nodes.length; i++) {
              if (that.nodes[i].ID == id) {
                that.nodes[i].Offline = 0
              }
            }
          } else if (8000 == response.data.code) {
            window.location.href="/ui/login.html"
          } else {
            alert(response.data.message);
          }
        }).catch(function (error) {
          alert(error);
        });
      },
      offline: function () {
        let id = $(event.target).attr("item-id");
        let that = this
        axios.post('/services/offline/' + id + '?time='+(new Date()).valueOf()).then(function (response) {
          console.log(response);
          if (2000 == response.data.code) {
            for (let i = 0; i < that.nodes.length; i++) {
              if (that.nodes[i].ID == id) {
                that.nodes[i].Offline = 1
              }
            }
          } else if (8000 == response.data.code) {
            window.location.href="/ui/login.html"
          } else {
            alert(response.data.message);
          }
        }).catch(function (error) {
          alert(error);
        });
      },
      del: function () {

      }
    }
  }
</script>
