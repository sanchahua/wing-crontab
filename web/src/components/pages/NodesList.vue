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
          <th class="sh-row">当前cpu负载</th>
          <th class="sh-row">当前内存使用率</th>
          <th class="sh-row">当前磁盘使用率</th>
          <th>操作</th>
        </tr>
        </thead>
        <tbody>
        <tr v-for="item in nodes">
          <th class="sh-row">hello</th>
          <td class="sh-row">127.0.0.1:38001</td>
          <td class="sh-row">online</td>
          <td class="sh-row">20%</td>
          <td class="sh-row">30%</td>
          <td class="sh-row">56%</td>
          <td>
            <div>
              <a class="btn" v-if="item.offline" v-on:click="online">上线</a>
              <a class="btn" v-else v-on:click="offline">下线</a>
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
    name: "NodesList",
    data: function() {
      return {
        nodes: [
          {
            offline: false,
          }
        ],
        datetime: (new Date()).Format("yyyy-MM-dd hh:mm:ss"),
      }
    },
    created: function() {
    },
    methods: {
      online: function () {

      },
      offline: function () {

      }
    }
  }
</script>
