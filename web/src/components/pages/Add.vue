<template>
  <!--/sub-heard-part-->
  <div>
  <div class="sub-heard-part">
    <ol class="breadcrumb m-b-0">
      <li><a href="/ui/#/">首页</a></li>
      <li class="active">增加定时任务</li>
    </ol>
  </div>
  <!--/sub-heard-part-->
  <!--/forms-->
  <div class="forms-main">
    <h2 class="inner-tittle">增加定时任务 </h2>
    <div class="graph-form">
      <div class="form-body">
        <div class="form-group">
          <label for="cron-blame">责任人</label>
          <!--<input type="text" class="form-control" id="cron-blame">-->
          <select class="form-control" id="cron-blame">
            <option v-for="item in users.data" v-bind:value="item.id">
              {{item.user_name}}<{{item.real_name}}>
            </option>
          </select>
        </div>
          <div class="form-group">
            <label for="cron-set">定时配置，如：*/1 * * * * *，这里精确到秒，前面的意思是每秒执行一次，分别对应，秒分时日月周</label>
            <input type="text" class="form-control" id="cron-set">
          </div>
          <div class="form-group">
            <label for="start-time">开始时间，大于等于此时间才执行，不限留空</label>
            <input type="text" class="form-control" id="start-time" v-bind:value="datetime">
          </div>
          <div class="form-group">
            <label for="end-time">结束时间，小于此时间才执行，不限留空</label>
            <input type="text" class="form-control" id="end-time" value="2099-01-01 08:00:00">
          </div>
          <div class="form-group">
            <label for="command">执行命令</label>
            <input type="text" class="form-control" id="command" value="">
          </div>
          <div class="checkbox">
            <label>
              <input type="checkbox" id="cron-stop"> 初始化为停止状态
            </label>
          </div>
          <div class="checkbox">
            <label>
              <input type="checkbox" id="cron-is-mutex"> 严格互斥执行
            </label>
          </div>
          <div class="form-group">
            <label for="remark">备注</label>
            <textarea class="form-control" id="remark"></textarea>
          </div>
          <button type="button" class="btn btn-default" id="do-submit">提交</button>
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
    name: "Add",
    data: function() {
      return {
        datetime: (new Date()).Format("yyyy-MM-dd hh:mm:ss"),
        users: {
          data: [],
        }
      }
    },
    created: function() {
      let script = document.createElement("script");
      script.src = "./static/js/add.vue.js?t=" +  (new Date()).valueOf();
      document.body.appendChild(script)
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
      }
    }
  }
</script>
