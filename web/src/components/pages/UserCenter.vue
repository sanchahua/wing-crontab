<template>
  <!--/sub-heard-part-->
  <div>
    <div class="sub-heard-part">
      <ol class="breadcrumb m-b-0">
        <li><a href="/ui/#/">首页</a></li>
        <li class="active">个人中心</li>
      </ol>
    </div>
    <h3 class="inner-tittle two">个人信息 <a style="color: #2e25e6; cursor: pointer; font-size: 12px; margin-left: 20px; text-decoration: underline;" class="showEdit" v-on:click="showEdit">编辑个人信息</a></h3>
      <div class="tables userinfo">
        <table class="table table-bordered">
          <tbody>
          <tr>
            <th scope="row">ID</th>
            <td>{{userinfo.id}}</td>
          </tr>
          <tr>
            <th scope="row">用户名</th>
            <td>{{userinfo.user_name}}</td>
          </tr>
          <tr>
            <th scope="row">真实姓名</th>
            <td>{{userinfo.real_name}}</td>
          </tr>
          <tr>
            <th scope="row">手机</th>
            <td>{{userinfo.phone}}</td>
          </tr>
          </tbody>
        </table>
      </div>

    <!--/sub-heard-part-->
    <!--/forms-->
    <div class="forms-main" style="display: none;">
      <h2 class="inner-tittle">编辑用户(*必填项) </h2>
      <div class="graph-form">
        <div class="form-body">
          <div class="form-group">
            <label for="user_name">*用户名</label>
            <input type="text" class="form-control" id="user_name" v-bind:value="userinfo.user_name" v-model="userinfo.user_name">
          </div>
          <div class="form-group">
            <label for="password">*密码</label>
            <input type="password" class="form-control" id="password" v-bind:value="userinfo.password" v-model="userinfo.password">
          </div>
          <div class="form-group">
            <label for="real_name">真实姓名</label>
            <input type="text" class="form-control" id="real_name" v-bind:value="userinfo.real_name" v-model="userinfo.real_name">
          </div>
          <div class="form-group">
            <label for="phone">手机</label>
            <input type="text" class="form-control" id="phone" v-bind:value="userinfo.phone" v-model="userinfo.phone">
          </div>

          <button type="button" class="btn btn-default" v-on:click="update">提交</button>
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
    name: "UserCenter",
    data: function() {
      return {
        userinfo: {
          id: 0,
          user_name: "",
          password: "",
          real_name: "",
          phone: "",
        }
      }
    },
    mounted: function() {
      this.getUserInfo()
    },
    methods: {
      getUserInfo: function() {
        let that = this;
        axios.get("/user/session/info").then(function (response) {
          console.log(response)
          if (2000 == response.data.code) {
            that.userinfo = response.data.data;
          } else if (8000 == response.data.code) {
            window.location.href="/ui/login.html"
          }
        }).catch(function (error) {
        });
      },
      showEdit: function(){
        $(".forms-main").toggle();
        $(".userinfo").toggle();
        if ($(".showEdit").html() == "编辑个人信息") {
          $(".showEdit").html("收起");
        } else {
          $(".showEdit").html("编辑个人信息")
        }
      },
      update: function () {
        let that = this
        if (that.userinfo.user_name == "") {
          return alert("用户名不能为空")
        }
        if (that.userinfo.password == "") {
          return alert("密码不能为空")
        }

        axios.post('/user/session/update?time='+(new Date()).valueOf(), {
          username: that.userinfo.user_name,
          password:  that.userinfo.password,
          real_name: that.userinfo.real_name,
          phone:  that.userinfo.phone,
        }).then(function (response) {
          console.log(response);
          if (2000 == response.data.code) {
            //window.location.href = "/ui/#/users"
            that.showEdit();
          } else if (8000 == response.data.code) {
            window.location.href="/ui/login.html"
          } else {
            alert(response.data.message);
          }
        }).catch(function (error) {

        });
      },
    }
  }
</script>
