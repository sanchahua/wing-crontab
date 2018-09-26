<template>
  <!--/sub-heard-part-->
  <div>
    <div class="sub-heard-part">
      <ol class="breadcrumb m-b-0">
        <li><a href="/ui/#/">首页</a></li>
        <li class="active">用户管理</li>
      </ol>
    </div>
    <!--/sub-heard-part-->
    <!--/forms-->
    <div class="forms-main">
      <h2 class="inner-tittle">增加用户(*必填项) </h2>
      <div class="graph-form">
        <div class="form-body">
          <div class="form-group">
            <label for="cron-blame">*用户名</label>
            <input type="text" class="form-control" id="cron-blame">
          </div>
          <div class="form-group">
            <label for="cron-set">*密码</label>
            <input type="text" class="form-control" id="cron-set">
          </div>
          <div class="form-group">
            <label for="start-time">真实姓名</label>
            <input type="text" class="form-control" id="start-time" v-bind:value="datetime">
          </div>
          <div class="form-group">
            <label for="end-time">手机</label>
            <input type="text" class="form-control" id="end-time" value="2099-01-01 08:00:00">
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
    name: "Users",
    data: function() {
      return {
        datetime: (new Date()).Format("yyyy-MM-dd hh:mm:ss"),
      }
    },
    created: function() {
      // let script = document.createElement("script");
      // script.src = "./static/js/add.vue.js?t=" +  (new Date()).valueOf();
      // document.body.appendChild(script)
    },
  }
</script>
