<template>
  <div>
    <!--<div><input type="text" v-model="logs.keyword"/><input type="button" value="搜索" v-on:click="search"/></div>-->
    <div>
      <label style="cursor: pointer;" v-on:click="searchFailLogs"><input v-model="logs.searchFail" type="checkbox"/>查看失败记录</label>
      <a v-on:click="prevPage" style="cursor: pointer;">上一页</a>
      <label>{{logs.page}}/{{logs.totalPage}}</label>
      <a v-on:click="nextPage" style="cursor: pointer;">下一页</a>
      自动刷新 <select v-on:change="refresh">
      <option value="0">不刷新</option>
      <option value="1" selected>1s</option>
      <option value="5">5s</option>
      <option value="10">10s</option>
      <option value="30">30s</option>
      <option value="60">60s</option>
    </select>
    </div>
    <table class="table table-bordered">
      <thead> <tr>
        <th>#Id</th>
        <th>定时任务Id</th>
        <th>进程Id</th>
        <th>开始执行时间</th>
        <th>结果</th>
        <th>耗时(毫秒)</th>
        <th>备注</th>
        <th>输出</th>
        <th>操作</th>
      </tr> </thead>
      <tbody>
      <tr v-for="item in logs.data">
        <th scope="row">{{item.id}}</th>
        <th scope="row">{{item.cron_id}}</th>
        <th scope="row">{{item.process_id}}</th>
        <td>{{item.start_time}}</td>
        <td>{{item.state}}</td>
        <td>{{item.use_time}}</td>
        <td style="word-break: break-all;">{{item.remark}}</td>
      <td style="word-break: break-all;">{{item.output}}</td>
        <td>
          <a class="btn" v-bind:item-process_id="item.process_id" v-bind:item-id="item.cron_id" v-on:click="kill">终止进程</a>
          <a class="btn" v-bind:item-id="item.id" v-on:click="detail">详情</a>
        </td>
      </tr>

        </tbody> </table>
  </div>
</template>
<script>
 // let intv = false
 let re = null
export default {
  name: "CronList",
  data: function () {
    return {
      logs: {
        data: [],
        limit: 50,
        page: 1,
        total: 0,
        totalPage: 0,
        searchFail: false,
        keyword: "",
      },
    }
  },
  mounted: function() {
    let oldhref = window.location.href;
    let that = this
    // if (!intv) {
    //   intv = true
    //   window.setInterval(function () {
    //     if (window.location.href != oldhref) {
    //       oldhref = window.location.href;
    //       that.getLogs("setInterval");
    //     }
    //   }, 5000);
    // }
    that.getLogs("mounted")
    re = window.setInterval(function () {
      that.getLogs("refresh setInterval")
    }, 1000)
  },
  methods:{
    searchFailLogs: function(){
      this.logs.page=1
      let that = this
      console.log(this.logs.searchFail)
      window.setTimeout(function () {
        console.log(that.logs.searchFail)
        that.getLogs("setTimeout")
      }, 20)
    },
    getLogs: function (callfrom) {
      console.log(callfrom)
      // /log/list/0/0/0
      let h = window.location.hash;
      let arr = h.split("?", -1)
      console.log(arr)
      let that = this;
      let params = {}
      if (arr.length > 1) {
        let pk = arr[1].split("&")
        let i = 0;
        let len = pk.length
        for (i = 0; i < len; i++) {
          let t = pk[i].split("=")
          if (t.length > 1) {
            params[t[0]] = t[1]
          }
        }
        console.log(params)
      }
      if (typeof params.id == "undefined") {
        params.id = 0;
      }
      let sf = "0"
      if (that.logs.searchFail) {
        sf = "1"
      }
      axios.get('/log/list/'+params.id+'/'+sf+'/'+that.logs.page+'/'+that.logs.limit+'?time='+(new Date()).valueOf()+"&keyword="+encodeURIComponent(that.logs.keyword)).then(function (response) {
        if (2000 == response.data.code) {
          console.log(response);
          that.logs.data = response.data.data.data
          that.logs.limit = response.data.data.limit
          that.logs.page = response.data.data.page
          that.logs.total = response.data.data.total
          that.logs.totalPage = response.data.data.totalPage
        } else {
          alert(response.data.message);
        }
      }).catch(function (error) {

      });
    },
    search: function() {
      this.getLogs("search")
    },
    prevPage: function () {
      let pregPage = this.logs.page-1
      if (pregPage < 1) {
        pregPage = this.logs.totalPage
      }
      this.logs.page = pregPage
      this.getLogs("prevPage")
    },
    nextPage: function () {
      let nextPage = this.logs.page+1
      if (nextPage > this.logs.totalPage) {
        nextPage = 1//that.logs.totalPage
      }
      this.logs.page = nextPage
      this.getLogs("nextPage")
    },
    refresh: function (ev) {
      let that = this
      let i = $(ev.target).val();
      console.log(i)
      if (re != null) {
        window.clearInterval(re)
      }
      if (i > 0) {
        re = window.setInterval(function () {
          that.getLogs("refresh setInterval")
        }, i * 1000)
      }
    },
    kill: function (event) {
      let id = $(event.target).attr("item-id");
      let process_id = $(event.target).attr("item-process_id");
      axios.get('/cron/kill/'+id+'/'+process_id+'?time='+(new Date()).valueOf()).then(function (response) {
        console.log(response)
        if (2000 == response.data.code) {
          $(event.target).html("kill成功")
        } else {
          $(event.target).html("kill失败")
        }
        window.setTimeout(function () {
          $(event.target).html("终止进程")
        }, 3000)
      }).catch(function (error) {
      });
    },
    detail: function (event) {
      let id = $(event.target).attr("item-id");
      window.location.href="/ui/#/log_detail?id="+id;
    }
  }
}
</script>
