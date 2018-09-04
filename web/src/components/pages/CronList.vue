<template>
  <div>
    <!--sub-heard-part-->
    <div class="sub-heard-part">
      <ol class="breadcrumb m-b-0">
        <li><a href="/ui/#/">首页</a></li>
        <li class="active">定时任务管理</li>
      </ol>
    </div>
    <!--//sub-heard-part-->
    <div class="graph-visual tables-main">

      <h3 class="inner-tittle two">定时任务列表（{{cron_list.length}}个） </h3>
      <div class="search-tool">
        <div><input type="text" v-model="keyword"/><input type="button" v-on:click="search_keyword" value="查询"/></div>
        <div v-on:click="quick_search">
        <div>
          <label><input name="is-stop" v-model="is_stop" value="2" type="radio"/>正在运行</label>
          <label><input name="is-stop" v-model="is_stop" value="1" type="radio"/>已停止</label>
          <label><input name="is-stop" v-model="is_stop" value="0" type="radio"/>取消</label>
        </div>
        <div>
          <label><input name="is-mutex" v-model="is_mutex" value="1" type="radio"/>互斥</label>
          <label><input name="is-mutex" v-model="is_mutex" value="2" type="radio"/>非互斥</label>
          <label><input name="is-mutex" v-model="is_mutex" value="0" type="radio"/>取消</label>
        </div>
        <div>
          <label><input name="is-timeout" v-model="is_timeout" value="1" type="radio"/>已过期</label>
          <label><input name="is-timeout" v-model="is_timeout" value="2" type="radio"/>有效期内</label>
          <label><input name="is-timeout" v-model="is_timeout" value="0" type="radio"/>取消</label>
        </div>
        </div>
      </div>
      <div>
        <div class="tables">
          <table class="table table-bordered" width="100%">
            <!--<thead>-->
              <!--<tr>-->
                <!--<th>#Id</th>-->
                <!--<th>定时配置</th>-->
                <!--<th>严格互斥</th>-->
                <!--<th>运行时间范围</th>-->
                <!--<th>执行指令</th>-->
                <!--<th>正在运行</th>-->
                <!--<th title="0-1属于正常，大于1说明有定时任务进程堆积">进程数<a style="color: #f00;">?</a></th>-->
                <!--<th>备注</th>-->
                <!--<th>操作</th>-->
              <!--</tr>-->
            <!--</thead>-->
            <tbody>
            <tr v-for="item in cron_list">
              <td scope="row">
                <div><span>id：</span><span>{{item.id}}</span></div>
                <div><span>定时配置：</span><span>{{item.cron_set}}</span></div>
                <div>
                  <span>互斥：</span>
                  <span v-if="item.is_mutex">是</span>
                  <span v-else>否</span>
                </div>
                <div>
                  <span>运行时间范围：</span>
                  <span>{{item.start_time}} - {{item.end_time}}</span>
                </div>
                <div>
                  <span>执行指令：</span>
                  <span>{{item.command}}</span>
                </div>
                <div>
                  <span>正在运行：</span>
                  <span v-if="item.stop"><label style="color: #f00; font-weight: bold;">否</label></span>
                  <span v-else>是</span>
                </div>
                <div><span>并行进程数：</span>
              <span>{{item.process_num}}</span>
                </div>
              <div>
                <span>备注信息：</span>
                <span>{{item.remark}}</span>
              </div>
              <div>
                <a class="btn" v-if="item.stop" v-bind:item-id="item.id" v-on:click="start">开始</a>
                <a class="btn" v-else v-bind:item-id="item.id" v-on:click="stop">停止</a>
                <a class="btn" v-bind:item-id="item.id" v-on:click="jumpEdit">编辑</a>
                <a class="btn" v-bind:item-id="item.id" v-on:click="del">删除</a>
                <a class="btn"  v-bind:item-id="item.id" v-if="item.is_mutex" v-on:click="mutex0">取消互斥</a>
                <a class="btn" v-else  v-bind:item-id="item.id" v-on:click="mutex1">设为互斥</a>
                <a class="btn" v-bind:item-id="item.id" v-on:click="jumpLogs">运行日志</a>
                <!--<a class="btn" v-bind:item-id="item.id" v-on:click="analysis">智能分析</a>-->
                <a class="btn" v-bind:item-id="item.id" v-bind:item-command="item.command" v-on:click="run">运行</a>
              </div>
              </td>
            </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
    <!--//graph-visual-->
  </div>
</template>
<script>
  export default {
    name: "CronList",
    data: function() {
      return {
        cron_list: [],
        is_stop: 0,
        is_mutex: 0,
        is_timeout: 0,
        keyword: "",
      }
    },
    mounted: function(){
      this.list();
    },
    methods: {
      list: function(){
          let that = this;
          let uri = '/cron/list?'
        if (that.is_stop == 0) {
         uri += 'stop=&'
        } else if(that.is_stop==1) {
          uri += 'stop=1&'
        } else if(that.is_stop==2) {
          uri += 'stop=0&'
        }

        if (that.is_mutex == 0) {
          uri += 'mutex=&'
        } else if(that.is_mutex==1) {
          uri += 'mutex=1&'
        } else if(that.is_mutex==2) {
          uri += 'mutex=0&'
        }

        if (that.is_timeout == 0) {
          uri += 'timeout=&'
        } else if(that.is_timeout==1) {
          uri += 'timeout=1&'
        } else if(that.is_timeout==2) {
          uri += 'timeout=0&'
        }
        uri += 'keyword=' + encodeURIComponent(that.keyword)
        console.log(uri)
          axios.get(uri).then(function (response) {
            if (2000 == response.data.code) {
              console.log(response.data.data);
              // let i=0;
              // for (i=0;i<response.data.data.length;i++) {
              //   if (response.data.data[i].start_time>0) {
              //     response.data.data[i].start_time = new Date(response.data.data[i].start_time*1000).Format("yyyy-MM-dd hh:mm:ss");
              //   } else {
              //     response.data.data[i].start_time = "不限";
              //   }
              //   if (response.data.data[i].end_time>0) {
              //     response.data.data[i].end_time = new Date(response.data.data[i].end_time*1000).Format("yyyy-MM-dd hh:mm:ss");
              //   } else {
              //     response.data.data[i].end_time = "不限";
              //   }
              // }
              that.cron_list = response.data.data
            } else {
              alert(response.data.message);
            }
          }).catch(function (error) {});
      },
      stop: function (event) {
          let id = $(event.target).attr("item-id");
          //alert("停止"+id);
          // /cron/stop/1656
          var that = this;
          axios.get('/cron/stop/'+id).then(function (response) {
            if (2000 == response.data.code) {
              console.log(response);
              let len = that.cron_list.length;// = response.data.data
              let i = 0;
              for (i = 0; i < len; i++) {
                if (that.cron_list[i].id == id) {
                  that.cron_list[i].stop = true;
                  console.log(that.cron_list[i]);
                }
              }
            } else {
              alert(response.data.message);
            }
          }).catch(function (error) {

          });
        },
      start: function (event) {
          let id = $(event.target).attr("item-id");
          var that = this;
          axios.get('/cron/start/'+id).then(function (response) {
            if (2000 == response.data.code) {
              console.log(response);
              let len = that.cron_list.length;// = response.data.data
              let i = 0;
              for (i = 0; i < len; i++) {
                if (that.cron_list[i].id == id) {
                  that.cron_list[i].stop = false;
                  console.log(that.cron_list[i]);
                }
              }
            } else {
              alert(response.data.message);
            }
          }).catch(function (error) {

          });
        },
      del: function (event) {
        let id = $(event.target).attr("item-id");
        if (!window.confirm("确认删除"+id + "？")) {
          return;
        }
        let that = this;
        axios.get('/cron/delete/'+id).then(function (response) {
          if (2000 == response.data.code) {
            console.log(response);
            //$(event.target).parents("tr").remove();
            let len = that.cron_list.length;// = response.data.data
            let i = 0;
            for (i = 0; i < len; i++) {
              if (that.cron_list[i].id == id) {
                console.log(that.cron_list[i]);
                that.cron_list.splice(i,1);
              }
            }
          } else {
            alert(response.data.message);
          }
        }).catch(function (error) {

        });
      },
      jumpEdit: function(event){
        let id = $(event.target).attr("item-id");
        window.location.href="/ui/#/edit?id=" + id
      },
      jumpLogs:function(){
        let id = $(event.target).attr("item-id");
        window.location.href="/ui/#/logs?id=" + id
      },
      // 取消互斥
      mutex0: function (event) {
        let id = $(event.target).attr("item-id");
        let that = this;
        axios.get('/cron/mutex/false/'+id).then(function (response) {
          if (2000 == response.data.code) {
            console.log(response);
            let len = that.cron_list.length;// = response.data.data
            let i = 0;
            for (i = 0; i < len; i++) {
              if (that.cron_list[i].id == id) {
                that.cron_list[i].is_mutex = false;
                console.log(that.cron_list[i]);
              }
            }
          } else {
            alert(response.data.message);
          }
        }).catch(function (error) {

        });
      },
      // 设为互斥
      mutex1: function (event) {
        let id = $(event.target).attr("item-id");
        let that = this;
        axios.get('/cron/mutex/true/'+id).then(function (response) {
          if (2000 == response.data.code) {
            console.log(response);
            let len = that.cron_list.length;// = response.data.data
            let i = 0;
            for (i = 0; i < len; i++) {
              if (that.cron_list[i].id == id) {
                that.cron_list[i].is_mutex = true;
                console.log(that.cron_list[i]);
              }
            }
          } else {
            alert(response.data.message);
          }
        }).catch(function (error) {

        });
      },
      analysis: function () {

      },
      run: function (event) {
        let id = $(event.target).attr("item-id");
        let command = $(event.target).attr("item-command");
        var d = dialog({
          title: '确认执行？',
          content:
          '<div style="color: #f00;">注意：即将执行指令['+command+']，最终产生的风险(系统风险、业务风险)请自行评估！</div>'+
          '<div>默认超时时间：<input id="property-returnValue-demo" value="3" />秒</div>'+
          '<div class="run-command-response">执行返回值：</div>'+
          '<div class="run-command-response"><textarea style="width: 100%;" id="run-command-response-show"></textarea></div>',
          ok: function () {
            let that = this
            that.title('执行中…');
            let value = $('#property-returnValue-demo').val();
            //alert(value)
            //this.close();
            //this.remove();

            axios.get('/cron/run/'+id +'/'+value).then(function (response) {
              if (2000 == response.data.code) {
                console.log(response);
                $("#run-command-response-show").val(response.data.data)
              } else {
                //alert(response.data.message);
                $("#run-command-response-show").val(response.data.message)
              }
              that.title('执行完成');
            }).catch(function (error) {
              $("#run-command-response-show").val("网络异常")
            });

            return false
          },
          okValue: '确定',
          cancelValue: '取消',
          cancel: function () {}
        });
        d.show();
      },
      quick_search: function (ev) {
        let that = this;
        window.setTimeout(function () {
          console.log("stop=", that.is_stop)
          console.log("is_mutex=", that.is_mutex)
          console.log("is_timeout=", that.is_timeout)
          console.log("keyword=", that.keyword)
          that.list()
        }, 500)
        ev.stopPropagation();
      },
      search_keyword: function () {
        let that = this;
        window.setTimeout(function () {
          that.list()
        }, 500)
      }

    },
    created: function () {
    }
  }

</script>
<style>
a.btn {
  border: 1px solid #2ecc71;
  padding: 2px 20px;
  font-size: 14px;
  margin: 2px 6px;
}
</style>
