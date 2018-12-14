<template>
  <div>
    <!--<div><input type="text" v-model="logs.keyword"/><input type="button" value="搜索" v-on:click="search"/></div>-->
    <div>
      开始时间 <input type="text" id="start-time" v-bind:value="logs.start_time">
      - 结束时间 <input type="text" id="end-time" v-bind:value="logs.end_time">
    </div>
    <div>
      定时任务id<input type="text" v-model="logs.cron_id"/> 搜索输出<input type="text" v-model="logs.output"/>
    </div>
    <div>运行时长>=<input type="text" v-model="logs.use_time"/>毫秒 <input type="button" value="搜索" v-on:click="search"/>
    </div>
    <div>
      <label style="cursor: pointer;" v-on:click="searchFailLogs">
        <input v-model="logs.searchFail" type="checkbox"/>查看失败记录
      </label>
      <label style="cursor: pointer;" v-on:click="searchResult">
        <input v-model="logs.searchResult" checked type="checkbox"/>只看结果
      </label>
      <label style="cursor: pointer;" v-on:click="hideRemarkAndOutput">
        <input checked type="checkbox"/>隐藏备注和输出
      </label>
      <a v-on:click="prevPage" style="cursor: pointer;">上一页</a>
      <label>{{logs.page}}/{{logs.totalPage}}</label>
      <a v-on:click="nextPage" style="cursor: pointer;">下一页</a>
      自动刷新 <select v-on:change="refresh">
      <option value="0" selected>不刷新</option>
      <option value="1">1s</option>
      <option value="5">5s</option>
      <option value="10">10s</option>
      <option value="30">30s</option>
      <option value="60">60s</option>
    </select>
    </div>
    <table class="table table-bordered">
      <thead> <tr>
        <th style="cursor: pointer;" v-on:click="sortbyid">#Id<img class="sort-tag" src="../../../static/images/sort.jpeg"/></th>
        <th>定时任务Id</th>
        <th>分发服务器</th>
        <th>运行服务器</th>
        <th>进程Id</th>
        <th style="cursor: pointer;" v-on:click="sortbystarttime">开始执行时间<img class="sort-tag" src="../../../static/images/sort.jpeg"/></th>
        <th>结果</th>
        <th style="cursor: pointer;" v-on:click="sortbyusetime">耗时(毫秒)<img class="sort-tag" src="../../../static/images/sort.jpeg"/></th>
        <th class="ro-text" style=" display: none;">备注</th>
        <th class="ro-text" style=" display: none;">输出</th>
        <th>操作</th>
      </tr> </thead>
      <tbody>
      <tr v-for="item in logs.data">
        <th scope="row">{{item.id}}</th>
        <th scope="row">{{item.cron_id}}</th>
        <th scope="row">{{item.dispatch_server_name}}</th>
        <th scope="row">{{item.run_server_name}}</th>
        <th scope="row">{{item.process_id}}</th>
        <td>{{item.start_time}}</td>
        <td>{{item.state}}</td>
        <td>{{item.use_time}}</td>
        <td class="ro-text" style="word-break: break-all; display: none;">{{item.remark}}</td>
        <td class="ro-text" style="word-break: break-all; display: none;">{{item.output}}</td>
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
        searchResult: true,
       // keyword: "",
        sort_by: "id",
        sort_type: "desc",
        start_time: "",
        end_time: "",
        cron_id: "",
        output: "",
        use_time: 0,
      },
    }
  },
  mounted: function() {
    let oldhref = window.location.href;
    let that = this
    //////////
    let h = window.location.hash;
    let arr = h.split("?", -1)
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
    if (params.id > 0) {
      that.logs.cron_id = params.id
    }
    that.getLogs("mounted")
    jeDate("#start-time",{
      festival:true,
      minDate:"1900-01-01",              //最小日期
      maxDate:"2099-12-31",              //最大日期
      method:{
        choose:function (params) {
          // alert(1)
        }
      },
      format: "YYYY-MM-DD hh:mm:ss",
      toggle: function(obj){
        // console.log(obj.val);      //得到日期生成的值，如：2017-06-16
        // alert(obj.val)
      },
      donefun: function(obj) {
        console.log(obj)
        that.logs.start_time = obj.val
      }
    });
    jeDate("#end-time",{
      festival:true,
      minDate:"1900-01-01",              //最小日期
      maxDate:"2099-12-31",              //最大日期
      method:{
        choose:function (params) {
          // console.log(params)
          // alert(1)
        }
      },
      format: "YYYY-MM-DD hh:mm:ss",
      toggle: function(obj){
        // console.log(obj.val);      //得到日期生成的值，如：2017-06-16
        // alert(obj.val)
      },
      donefun: function(obj) {
        console.log(obj)
        that.logs.end_time = obj.val
      }
    });
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
    searchResult: function() {
      this.logs.page=1
      let that = this
      console.log(this.logs.searchResult)
      window.setTimeout(function () {
        console.log(that.logs.searchResult)
        that.getLogs("setTimeout")
      }, 20)
    },
    sortbyid: function(event) {
      this.logs.page=1
      this.logs.sort_by = "id"
      if (this.logs.sort_type == "asc") {
        this.logs.sort_type = "desc"
      } else {
        this.logs.sort_type = "asc"
      }
      this.getLogs("sort")
    },
    sortbyusetime: function(event) {
      this.logs.page=1
      this.logs.sort_by = "use_time"
      if (this.logs.sort_type == "asc") {
        this.logs.sort_type = "desc"
      } else {
        this.logs.sort_type = "asc"
      }
      this.getLogs("sort")
    },
    sortbystarttime: function(){
      this.logs.page=1
      this.logs.sort_by = "start_time"
      if (this.logs.sort_type == "asc") {
        this.logs.sort_type = "desc"
      } else {
        this.logs.sort_type = "asc"
      }
      this.getLogs("sort")
    },
    getLogs: function (callfrom) {
      console.log(callfrom)
      let that = this;
      let sf = "0"
      if (that.logs.searchFail) {
        sf = "1"
      }
      let searchResult = 1
      console.log(that.logs.searchResult)
      if (!that.logs.searchResult) {
        searchResult = 0
      }
      console.log(that.logs.cron_id)

      let scronid  = that.logs.cron_id
      if (scronid == "") {
        scronid = 0
      }

      axios.get('/log/list/'+scronid+'/'+sf+'/'+that.logs.page+'/'+that.logs.limit+
        '?time='+(new Date()).valueOf()+
        // "&keyword="+encodeURIComponent(that.logs.keyword) +
        "&search_result="+searchResult+
        "&sort_by=" + that.logs.sort_by+
        "&sort_type=" + that.logs.sort_type+
        "&start_time=" + encodeURIComponent(that.logs.start_time) +
        "&end_time=" + encodeURIComponent(that.logs.end_time) +
        "&output=" + encodeURIComponent(that.logs.output) +
        "&use_time=" + that.logs.use_time
      ).then(function (response) {
        if (2000 == response.data.code) {
          console.log(response);
          that.logs.data = response.data.data.data
          that.logs.limit = response.data.data.limit
          that.logs.page = response.data.data.page
          that.logs.total = response.data.data.total
          that.logs.totalPage = response.data.data.totalPage
        } else if (8000 == response.data.code) {
          window.location.href="/ui/login.html"
        } else {
          console.log(response.data.message);
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
        } else if (8000 == response.data.code) {
          window.location.href="/ui/login.html"
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
    },
    hideRemarkAndOutput: function (event) {
      if ($(event.target).prop("checked")) {
        $(".ro-text").hide()
      } else {
        $(".ro-text").show()
      }
    }
  }
}
</script>
<style>
  .sort-tag{
    width: 20px;
  }
</style>
