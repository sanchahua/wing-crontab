<template>
  <!--/sub-heard-part-->
  <div>
    <div class="sub-heard-part">
      <ol class="breadcrumb m-b-0">
        <li><a href="/ui/#/">首页</a></li>
        <li class="active">增加用户</li>
      </ol>
    </div>
    <!--/sub-heard-part-->
    <!--/forms-->
    <div class="forms-main">
      <h2 class="inner-tittle">增加用户(*必填项) </h2>
      <div class="graph-form">
        <div class="form-body">
          <div class="form-group">
            <label for="user_name">*用户名</label>
            <input type="text" class="form-control" id="user_name" v-model="user_name">
          </div>
          <div class="form-group">
            <label for="password">*密码</label>
            <input type="text" class="form-control" id="password" v-model="password">
          </div>
          <div class="form-group">
            <label for="real_name">真实姓名</label>
            <input type="text" class="form-control" id="real_name" v-model="real_name">
          </div>
          <div class="form-group">
            <label for="phone">手机</label>
            <input type="text" class="form-control" id="phone" v-model="phone">
          </div>

          <button type="button" class="btn btn-default" v-on:click="add">提交</button>
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
    name: "UserAdd",
    data: function() {
      return {
        user_name: "",
        password: "",
        real_name: "",
        phone: "",
        datetime: (new Date()).Format("yyyy-MM-dd hh:mm:ss"),
      }
    },
    created: function() {
      // let script = document.createElement("script");
      // script.src = "./static/js/add.vue.js?t=" +  (new Date()).valueOf();
      // document.body.appendChild(script)
    },
    methods: {
      add: function () {
        let that = this
        if (that.user_name == "") {
          return alert("用户名不能为空")
        }
        if (that.password == "") {
          return alert("密码不能为空")
        }
        axios.post('/user/register?time='+(new Date()).valueOf(), {
          user_name: that.user_name,
          password:  that.password,
          real_name: that.real_name,
          phone:  that.phone,
        }).then(function (response) {
          if (2000 == response.data.code) {
            console.log(response);
            window.location.href = "/ui/#/users"
          } else {
            alert(response.data.message);
          }
        }).catch(function (error) {

        });
      },
    }
  }
</script>
