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
    <h3 class="inner-tittle two">用户列表</h3>
    <div class="graph">
      <div class="tables">
        <table class="table table-bordered"> <thead>
        <tr> <th>ID</th> <th>用户名</th> <th>状态</th> <th>真实姓名</th> <th>手机号码</th>
          <th>添加时间</th> <th>最后更新</th>
          <th>操作</th>
        </tr> </thead> <tbody>
        <tr v-for="item in users.data">
          <th scope="row">{{item.id}}</th>
          <td>{{item.user_name}}</td>
          <td v-if="item.enable">启用</td>
          <td v-else>禁用</td>

          <td>{{item.real_name}}</td> <td>{{item.phone}}</td><td>{{item.created}}</td>
          <td>{{item.updated}}</td>
          <td>
            <a class="bth" style="cursor: pointer;" v-if="item.enable" v-bind:data-id="item.id" data-enable="0" v-on:click="enable">禁用</a>
            <a class="bth" style="cursor: pointer;" v-else v-bind:data-id="item.id" data-enable="1" v-on:click="enable">启用</a>
            <a class="bth" style="cursor: pointer;" v-bind:data-id="item.id" v-on:click="jumpEdit">编辑</a>
          </td>
        </tr>
        </tbody> </table>
      </div>

    </div>

  </div>
  <!--//forms-->
</template>
<script>
  export default {
    name: "Users",
    data: function() {
      return {
        users: {
          data: [],
        }
      }
    },
    mounted: function() {
      this.getList()
    },
    methods:  {
      getList: function () {
        var that = this
        axios.get('/users?time='+(new Date()).valueOf()).then(function (response) {
          if (2000 == response.data.code) {
            that.users.data = response.data.data
            console.log(response);
          } else {
            alert(response.data.message);
          }
        }).catch(function (error) {

        });
      },
      jumpEdit:  function(event) {
        let id=$(event.target).attr("data-id")
        window.location.href="/ui/#/user_edit?id="+id
      },
      enable: function(event) {
        let that = this
        let id=$(event.target).attr("data-id")
        let enable=$(event.target).attr("data-enable")
        // /user/enable/{id}/{enable}
        axios.post('/user/enable/'+id+'/' + enable +'?time='+(new Date()).valueOf()).then(function (response) {
          console.log(response);
          if (2000 == response.data.code) {
            let len = that.users.data.length;
            for (let i=0;i<len;i++) {
              if (that.users.data[i].id == id) {
                if (enable == "1") {
                  that.users.data[i].enable = true
                } else {
                  that.users.data[i].enable = false
                }
                break
              }
            }

          } else {
            alert(response.data.message);
          }
        }).catch(function (error) {
          alert(error);
        });
      }
    }
  }
</script>
