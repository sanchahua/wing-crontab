<template>
  <!--/sub-heard-part-->
  <div>
    <div class="sub-heard-part">
      <ol class="breadcrumb m-b-0">
        <li><a href="/ui/#/">首页</a></li>
        <li class="active">编辑用户</li>
      </ol>
    </div>
    <!--/sub-heard-part-->
    <!--/forms-->
    <div class="forms-main">
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
    name: "UserEdit",
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
        if (params.id <= 0) {
          return;
        }

        axios.get("/user/info/"+params.id).then(function (response) {
          console.log(response)
          if (2000 == response.data.code) {
            that.userinfo = response.data.data;
          } else if (8000 == response.data.code) {
            window.location.href="/ui/login.html"
          }
        }).catch(function (error) {
        });
      },
      update: function () {
        let that = this
        if (that.userinfo.user_name == "") {
          return alert("用户名不能为空")
        }
        if (that.userinfo.password == "") {
          return alert("密码不能为空")
        }

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
        if (typeof params.id == "undefined") {
          params.id = 0;
        }
        if (params.id <= 0) {
          return;
        }

        axios.post('/user/update/'+params.id+'?time='+(new Date()).valueOf(), {
          username: that.userinfo.user_name,
          password:  that.userinfo.password,
          real_name: that.userinfo.real_name,
          phone:  that.userinfo.phone,
        }).then(function (response) {
          console.log(response);
          if (2000 == response.data.code) {
            window.location.href = "/ui/#/users"
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
