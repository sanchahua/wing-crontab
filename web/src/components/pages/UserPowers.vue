<template>
  <!--/sub-heard-part-->
  <div>
    <div class="sub-heard-part">
      <ol class="breadcrumb m-b-0">
        <li><a href="/ui/#/">首页</a></li>
        <li class="active">权限管理</li>
      </ol>
    </div>
    <!--/sub-heard-part-->
    <h3 class="inner-tittle two"><{{userinfo.user_name}}>的权限</h3>
    <div v-if="userinfo.admin" style="font-size: 14px; padding-left: 8px; border-left: #f00 8px solid; color: #f00;"><{{userinfo.user_name}}>是管理员，具有所有的权限，无需以下设置</div>
    <div class="graph">
      <div class="tables">
        <div><label style="margin-left: 15px;" v-on:click="checkAll"><input class="selected-all" type="checkbox"/>全选</label><a class="btn bth" v-on:click="save">保存</a></div>
        <table class="table table-bordered">
          <tbody>
          <tr v-for="item in powers">
            <td>
              <input v-bind:data-id="item.id" v-if="item.checked" checked class="power-row" v-bind:id="'row-'+item.id" type="checkbox"/>
              <input v-bind:data-id="item.id" v-else class="power-row" v-bind:id="'row-'+item.id" type="checkbox"/>
            </td>
            <td><label v-bind:for="'row-'+item.id">[{{item.id}}]{{item.name}}</label></td>
          </tr>
        </tbody>
        </table>
        <div><label style="margin-left: 15px;" v-on:click="checkAll"><input class="selected-all" type="checkbox"/>全选</label><a class="btn bth" v-on:click="save">保存</a></div>
      </div>

    </div>

  </div>
  <!--//forms-->
</template>
<script>
  export default {
    name: "UserPowers",
    data: function() {
      return {
        userinfo: {
          user_name: "",
        },
        powers: [{
          checked: false,
          id: 0,
          name: "",
        }]
      }
    },
    mounted: function() {
      //this.getPowers()
      this.getUserInfo()
    },
    methods:  {
      getParams: function() {
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
        }
        if (typeof params.id == "undefined") {
          params.id = 0;
        }
        return params;
      },
      save: function() {
        let powers = 0;
        $(".power-row").each(function() {
          if ($(this).prop("checked")) {
            let id = $(this).attr("data-id")
            id = parseInt(id)
            powers |= id
          }
        });
        console.log(powers)
        let that = this
        let params = that.getParams()
        axios.post("/user/powers/"+params.id+"/"+powers).then(function (response) {
          console.log(response)
          if (2000 != response.data.code) {
            alert(response.data.message);
          } else if (8000 == response.data.code) {
            window.location.href="/ui/login.html"
          } else {
            alert("保存成功");
          }
        }).catch(function (error) {
        });
      },
      checkAll: function(event) {
        $(".power-row").prop("checked", $(event.target).prop("checked"))
      },
      getUserInfo: function() {
        let that = this
        let params = that.getParams()
        axios.get("/user/info/"+params.id).then(function (response) {
          console.log(response)
          if (2000 == response.data.code) {
            that.userinfo = response.data.data;
          } else if (8000 == response.data.code) {
            window.location.href="/ui/login.html"
          }
          that.getPowers()
        }).catch(function (error) {
        });
      },
      getPowers: function () {
        let that = this
        axios.get('/powers?time='+(new Date()).valueOf()).then(function (response) {
          if (2000 == response.data.code) {
            that.powers = response.data.data
            let i = 0;
            let powers = parseInt(that.userinfo.powers)
            for (i = 0; i < that.powers.length; i++) {
              let id = parseInt(that.powers[i].id)
              that.powers[i].checked = ((id & powers) > 0)
            }
            window.setTimeout(function () {
              let selected = 0;
              $(".power-row").each(function() {
                if ($(this).prop("checked")) {
                  selected++
                }
              });
              if (selected == $(".power-row").length) {
                $(".selected-all").prop("checked", true)
              }
            }, 300)
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
