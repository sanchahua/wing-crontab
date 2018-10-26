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
          <!--<th class="sh-row">当前cpu负载</th>
          <th class="sh-row">当前内存使用率</th>
          <th class="sh-row">当前磁盘使用率</th>-->
          <th>操作</th>
        </tr>
        </thead>
        <tbody>
        <tr v-for="item in nodes">
          <th class="sh-row">{{item.Name}}</th>
          <td class="sh-row">{{item.Address}}</td>

          <td class="sh-row" v-if="item.Status">在线</td>
          <td class="sh-row" v-else>离线</td>

          <td class="sh-row" v-if="item.Leader">是</td>
          <td class="sh-row" v-else>否</td>

          <!--<td class="sh-row">20%</td>
          <td class="sh-row">30%</td>
          <td class="sh-row">56%</td>-->
          <td>
            <div>
              <a class="btn" v-if="item.Status" v-on:click="offline">下线</a>
              <a class="btn" v-else v-on:click="online">上线</a>
              <a class="btn" v-on:click="del">删除</a>
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

      },
      offline: function () {

      },
      del: function () {

      }
    }
  }
</script>
