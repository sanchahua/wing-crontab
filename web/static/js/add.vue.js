
$(document).ready(function(){
  jeDate("#start-time",{
    festival:true,
    minDate:"1900-01-01",              //最小日期
    maxDate:"2099-12-31",              //最大日期
    method:{
      choose:function (params) {

      }
    },
    format: "YYYY-MM-DD hh:mm:ss"
  });
  jeDate("#end-time",{
    festival:true,
    minDate:"1900-01-01",              //最小日期
    maxDate:"2099-12-31",              //最大日期
    method:{
      choose:function (params) {

      }
    },
    format: "YYYY-MM-DD hh:mm:ss"
  });

  $("#do-submit").click(function(){
    var blame    = $("#cron-blame").val();
    var stop     = $("#cron-stop").prop("checked")?"1":"0";
    var is_mutex = $("#cron-is-mutex").prop("checked")?"1":"0";
    var data = {
      cron_set:   $("#cron-set").val(),
      command:    $("#command").val(),
      start_time: $("#start-time").val(),
      end_time:   $("#end-time").val(),
      remark:     $("#remark").val(),
      stop:       stop,
      is_mutex:   is_mutex,
      blame:      blame,
    };
    // axios.defaults.headers.post['Content-Type'] = 'application/x-www-form-urlencoded';
    // JSON.stringify
    axios.post('/cron/add', data).then(function (response) {
        console.log(response);
        if (2000 == response.data.code) {
          // 转到管理页面
          window.location.href="/ui/#/cron_list";
        } else {
          alert(response.data.message);
        }
    }).catch(function (error) {
        alert(error);
    });
  });
})
