<template>
  <!--/sub-heard-part-->
  <div>
    <div><span class="tdis">定时任务ID</span><span>{{log_detail.cron_id}}</span></div>
    <div><span class="tdis">进程ID</span><span>{{log_detail.process_id}}</span></div>
    <div><span class="tdis">状态</span><span>{{log_detail.state}}</span></div>
    <div><span class="tdis">开始时间</span><span>{{log_detail.start_time}}</span></div>
    <div><span class="tdis">耗时(毫秒)</span><span>{{log_detail.use_time}}</span></div>
    <div><span class="tdis">备注</span><span>{{log_detail.remark}}</span></div>
    <div><span class="tdis">输出</span><span>{{log_detail.output}}</span></div>
  </div>
</template>
<script>
  export default {
    name: "LogDetail",
    data: function () {
      return {
        log_detail: {
          id: "",
          cron_id: "",
          process_id: "",
          state: "",
          start_time: "",
          use_time: 0,
          remark: "",
          output: "",
        },
      }
    },
    methods: {
      getInfo: function() {
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

        axios.get('/cron/log/detail/'+params.id).then(function (response) {
          console.log(response)
          if (2000 == response.data.code) {
            that.log_detail = response.data.data;
            console.log(that.log )
          } else if (8000 == response.data.code) {
            window.location.href="/ui/login.html"
          }
        }).catch(function (error) {
        });
      },
    },
    mounted: function () {
      this.getInfo()
    }
  }
</script>
<style>
  .tdis{
    display: inline-block;
    width: 120px;
    padding-right: 20px;
  }
</style>
