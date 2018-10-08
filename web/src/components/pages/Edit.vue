<template>
  <!--/sub-heard-part-->
  <div>
    <div class="sub-heard-part">
      <ol class="breadcrumb m-b-0">
        <li><a href="/ui/#/">首页</a></li>
        <li class="active">编辑定时任务</li>
      </ol>
    </div>
    <!--/sub-heard-part-->
    <!--/forms-->
    <div class="forms-main">
      <h2 class="inner-tittle">编辑定时任务 </h2>
      <div class="graph-form">
        <div class="form-body">
          <div class="form-group">
            <label for="cron-blame">责任人</label>
            <!--<input type="text" class="form-control" id="cron-blame" v-bind:value="cron_info.blame" v-model="cron_info.blame">-->
            <select class="form-control" id="cron-blame" v-model="cron_info.blame">
              <option v-for="item in users.data" v-bind:value="item.id">
                {{item.user_name}}<{{item.real_name}}>
              </option>
            </select>
          </div>
          <div class="form-group">
            <label for="cron-set">定时配置，如：*/1 * * * * *，这里精确到秒，前面的意思是每秒执行一次，分别对应，秒分时日月周</label>
            <input type="text" class="form-control" id="cron-set" v-bind:value="cron_info.cron_set" v-model="cron_info.cron_set">
          </div>
          <div class="form-group">
            <label>开始时间，大于等于此时间才执行，不限留空</label>
            <input type="text" class="form-control" id="start-time" v-bind:value="cron_info.start_time" v-model="cron_info.start_time">
          </div>
          <div class="form-group">
            <label for="end-time">结束时间，小于此时间才执行，不限留空</label>
            <input type="text" class="form-control" id="end-time" v-bind:value="cron_info.end_time" v-model="cron_info.end_time">
          </div>
          <div class="form-group">
            <label for="command">执行命令</label>
            <input type="text" class="form-control" id="command" v-bind:value="cron_info.command" v-model="cron_info.command">
          </div>
          <div class="checkbox">
            <label>
              <input type="checkbox" id="cron-stop" v-if="cron_info.stop" checked v-model="cron_info.stop">
              <input type="checkbox" id="cron-stop" v-else v-model="cron_info.stop">
              初始化为停止状态
            </label>
          </div>
          <div class="checkbox">
            <label>
              <input type="checkbox" id="cron-is-mutex" v-if="cron_info.is_mutex" checked v-model="cron_info.is_mutex">
              <input type="checkbox" id="cron-is-mutex" v-else v-model="cron_info.is_mutex">
              严格互斥执行
            </label>
          </div>
          <div class="form-group">
            <label for="remark">备注</label>
            <textarea class="form-control" id="remark" v-model="cron_info.remark">{{cron_info.remark}}</textarea>
          </div>

          <button type="button" class="btn btn-default" id="do-submit" v-on:click="submit">提交</button>
        </div>

      </div>
      <!--/forms-inner-->
      <!--//forms-inner-->
    </div>
  </div>
  <!--//forms-->
</template>
<script>



  export default {
    name: "Edit",
    data: function(){
      return {
        cron_info: {
          cron_set: "",
        },
        users: {
          data: [],
        }
      }
    },
    mounted: function(){
      let that = this;
      // window.setInterval(function () {
      //   console.log(that.cron_info)
      // }, 1000);
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
          that.cron_info.start_time = obj.val
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
          that.cron_info.end_time = obj.val
        }
      });
      this.getInfo()
      this.getUsers()
    },

    methods: {
      getUsers: function () {
        let that = this
        axios.get('/users?time='+(new Date()).valueOf()).then(function (response) {
          if (2000 == response.data.code) {
            that.users.data = response.data.data
            console.log(response);
          } else if (8000 == response.data.code) {
            window.location.href="/ui/login.html"
          } else {
            alert(response.data.message);
          }
        }).catch(function (error) {

        });
      },
      getInfo: function () {
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
          return;
        }
        axios.get('/cron/info/'+params.id+'?time='+(new Date()).valueOf()).then(function (response) {
          if (2000 == response.data.code) {
            console.log(response);
            that.cron_info = response.data.data
            // if (that.cron_info.start_time > 0) {
            //   that.cron_info.start_time = new Date(that.cron_info.start_time*1000).Format("yyyy-MM-dd hh:mm:ss");
            // } else {
            //   that.cron_info.start_time = "";
            // }
            // if (that.cron_info.end_time > 0) {
            //   that.cron_info.end_time = new Date(that.cron_info.end_time*1000).Format("yyyy-MM-dd hh:mm:ss");
            // } else {
            //   that.cron_info.end_time = "";
            // }
          } else if (8000 == response.data.code) {
            window.location.href="/ui/login.html"
          } else {
            alert(response.data.message);
          }
        }).catch(function (error) {

        });
      },
      submit: function () {
        // var stop     = $("#cron-stop").prop("checked")?"1":"0";
        // var is_mutex = $("#cron-is-mutex").prop("checked")?"1":"0";

        let data = {
          cron_set:   this.cron_info.cron_set,
          command:    this.cron_info.command,
          // vue 对于js动态修改intput的值，没办法通过v-model绑定获取
          start_time: $("#start-time").val(),
          end_time:   $("#end-time").val(),
          remark:     this.cron_info.remark,
          stop:       this.cron_info.stop?1:0,
          is_mutex:   this.cron_info.is_mutex?1:0,
          blame:      this.cron_info.blame,
        };
        // axios.defaults.headers.post['Content-Type'] = 'application/x-www-form-urlencoded';
        // JSON.stringify
        console.log(this.cron_info, data)
        let h = window.location.hash;
        let arr = h.split("?", -1)
        console.log(arr)
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
        let that = this
        axios.post('/cron/update/' + params.id+'?time='+(new Date()).valueOf(), data).then(function (response) {
          console.log(response);
          if (2000 == response.data.code) {
            // 转到管理页面
            window.location.href="/ui/#/cron_list";
          } else if (8000 == response.data.code) {
            window.location.href="/ui/login.html"
          } else {
            alert(response.data.message);
          }
        }).catch(function (error) {
          alert(error);
        });
      }
    },

    created: function() {
      // let script = document.createElement("script");
      // script.src = "./static/js/add.vue.js?t=" +  (new Date()).valueOf();
      // document.body.appendChild(script)
    },
  }
</script>
