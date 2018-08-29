
$(document).ready(function(){
  // var enLang = {
  //   name  : "en",
  //   month : ["01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "11", "12"],
  //   weeks : [ "SUN","MON","TUR","WED","THU","FRI","SAT" ],
  //   times : ["Hour","Minute","Second"],
  //   timetxt: ["Time","Start Time","End Time"],
  //   backtxt:"Back",
  //   clear : "Clear",
  //   today : "Now",
  //   yes   : "Confirm",
  //   close : "Close"
  // }
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
    var stop = $("#cron-stop").prop("checked")?"1":"0";
    var is_mutex = $("#cron-is-mutex").prop("checked")?"1":"0";
    var data = {
      cron_set: $("#cron-set").val(),
      command: $("#command").val(),
      start_time: $("#start-time").val(),
      end_time: $("#end-time").val(),
      remark: $("#remark").val(),
      stop: stop,
      is_mutex: is_mutex,
    };
    // axios.defaults.headers.post['Content-Type'] = 'application/x-www-form-urlencoded';
    // JSON.stringify
    axios.post('/cron/add', JSON.stringify(data)).then(function (response) {
        console.log(response);
    }).catch(function (error) {
        console.log(error);
    });
  });
})
