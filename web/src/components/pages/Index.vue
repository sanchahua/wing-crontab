<template>
  <!--custom-widgets-->
  <div>
    <div class="custom-widgets" author="yuyi">
      <div class="row-one">
        <div class="col-md-3 widget">
          <div class="stats-left ">
            <h5>定时任务</h5>
            <h4> 总数</h4>
          </div>
          <div class="stats-right">
            <label>{{cron_count}}</label>
          </div>
          <div class="clearfix"> </div>
        </div>
        <div class="col-md-3 widget states-mdl">
          <div class="stats-left">
            <h5>历史执行</h5>
            <h4>次数</h4>
          </div>
          <div class="stats-right">
            <label v-html="history_run_count">{{history_run_count}}</label>
          </div>
          <div class="clearfix"> </div>
        </div>
        <div class="col-md-3 widget states-thrd">
          <div class="stats-left">
            <h5>今日执行</h5>
            <h4>次数</h4>
          </div>
          <div class="stats-right">
            <label v-html="day_run_count">{{day_run_count}}</label>
          </div>
          <div class="clearfix"> </div>
        </div>
        <div class="col-md-3 widget states-last">
          <div class="stats-left">
            <h5>今日错误</h5>
            <h4>次数</h4>
          </div>
          <div class="stats-right">
            <label v-html="day_run_fail_count">{{day_run_fail_count}}</label>
          </div>
          <div class="clearfix"> </div>
        </div>
        <div class="clearfix"> </div>
      </div>
    </div>
    <!--//custom-widgets-->
    <!--/candile-->
    <!--<div class="candile">-->
      <!--<div class="candile-inner">-->
        <!--<h3 class="sub-tittle">Candlestick Chart </h3>-->
        <!--<div id="center"><div id="fig">-->
          <!--<script type="text/javascript+protovis">-->

															<!--/* Parse dates. */-->
															<!--var dateFormat = pv.Format.date("%d-%b-%y");-->
															<!--vix.forEach(function(d) d.date = dateFormat.parse(d.date));-->

															<!--/* Scales. */-->
															<!--var w =1220,-->
																<!--h = 300,-->
																<!--x = pv.Scale.linear(vix, function(d) d.date).range(0, w),-->
																<!--y = pv.Scale.linear(vix, function(d) d.low, function(d) d.high).range(0, h).nice();-->

															<!--var vis = new pv.Panel()-->
																<!--.width(w)-->
																<!--.height(h)-->
																<!--.margin(10)-->
																<!--.left(30);-->

															<!--/* Dates. */-->
															<!--vis.add(pv.Rule)-->
																 <!--.data(x.ticks())-->
																 <!--.left(x)-->
																 <!--.strokeStyle("#eee")-->
															   <!--.anchor("bottom").add(pv.Label)-->
																 <!--.text(x.tickFormat);-->

															<!--/* Prices. */-->
															<!--vis.add(pv.Rule)-->
																 <!--.data(y.ticks(7))-->
																 <!--.bottom(y)-->
																 <!--.left(-10)-->
																 <!--.right(-10)-->
																 <!--.strokeStyle(function(d) d % 10 ? "#ddd" : "#ddd")-->
															   <!--.anchor("left").add(pv.Label)-->
																 <!--.textStyle(function(d) d % 10 ? "#999" : "#ddd")-->
																 <!--.text(y.tickFormat);-->

															<!--/* Candlestick. */-->
															<!--vis.add(pv.Rule)-->
																<!--.data(vix)-->
																<!--.left(function(d) x(d.date))-->
																<!--.bottom(function(d) y(Math.min(d.high, d.low)))-->
																<!--.height(function(d) Math.abs(y(d.high) - y(d.low)))-->
																<!--.strokeStyle(function(d) d.open < d.close ? "#052963" : "#00C6D7")-->
															  <!--.add(pv.Rule)-->
																<!--.bottom(function(d) y(Math.min(d.open, d.close)))-->
																<!--.height(function(d) Math.abs(y(d.open) - y(d.close)))-->
																<!--.lineWidth(10);-->

															<!--vis.render();-->

																<!--</script>-->

        <!--</div>-->
        <!--</div>-->

      <!--</div>-->

    <!--</div>-->
    <!--/candile-->

    <!--/charts-->
    <div class="charts">
      <div class="chrt-inner">
        <div class="chrt-bars">
          <!--<div class="col-md-6 chrt-two">-->
            <!--<h3 class="sub-tittle">Bar Chart </h3>-->
            <!--<div id="chart2"></div>-->

          <!--</div>-->
          <div class="candile-inner">
            <h3 class="sub-tittle sub-tittle-e">执行次数 </h3>
            <span class="time-tool">
              <label><input name="time-pass" value="7" v-model="days" type="radio" checked/>过去一周</label>
              <label><input name="time-pass" value="30" v-model="days" type="radio"/>过去一个月</label>
              <label><input name="time-pass" value="90" v-model="days" type="radio"/>过去三个月</label>
              <label><input name="time-pass" value="180" v-model="days" type="radio"/>过去六个月</label>
              <label><input name="time-pass" value="365" v-model="days" type="radio"/>过去一年</label>
            </span>
            <div id="chartdiv"></div>
          </div>
          <div class="clearfix"> </div>
        </div>

      </div>
      <!--/charts-inner-->
    </div>
  </div>
  <!--//outer-wp-->
</template>
<script>
  export default {
    name: "Index",
    data: function(){
      return {
        days: 7,
        cron_count: 0,
        history_run_count: 0,
        day_run_count: 0,
        day_run_fail_count: 0,
      }
    },
    methods: {
      getStatistics: function () {
        let that = this;
        axios.get('/index').then(function (response) {
          if (2000 == response.data.code) {
            console.log(response);
            that.cron_count         = response.data.data.cron_count
            that.history_run_count  = response.data.data.history_run_count
            that.day_run_count      = response.data.data.day_run_count
            that.day_run_fail_count = response.data.data.day_run_fail_count

            if (that.history_run_count >= 10000) {
              that.history_run_count = (Math.floor(that.history_run_count/10000*100)/100) + "<a style='font-size: 12px; font-weight: bold; color: #000;'>万</a>"
            }

            if (that.day_run_count >= 10000) {
              that.day_run_count = (Math.floor(that.day_run_count/10000*100)/100) + "<a style='font-size: 12px; font-weight: bold; color: #000;'>万</a>"
            }

            if (that.day_run_fail_count >= 10000) {
              that.day_run_fail_count = (Math.floor(that.day_run_fail_count/10000*100)/100) + "<a style='font-size: 12px; font-weight: bold; color: #000;'>万</a>"
            }

          } else {
            alert(response.data.message);
          }
        }).catch(function (error) {

        });
      },
      getCharts: function (chart) {
        let that = this
        axios.get('/charts/'+that.days).then(function (response) {
          console.log(response)
          if (2000 == response.data.code) {
            chart.dataProvider = response.data.data;
            chart.validateNow();
            chart.validateData()
          } else {
            alert(response.data.message);
          }
        }).catch(function (error) {

        });
      },
    },
    mounted: function(){
      let that = this;


      let chart = AmCharts.makeChart("chartdiv", {
        "type": "serial",
        "theme": "light",
        "dataDateFormat": "YYYY-MM-DD",
        "legend": {
          "useGraphSettings": true
        },
        "dataProvider": [],
        "synchronizeGrid":true,
        "valueAxes": [{
          "id":"v1",
          "axisColor": "#FF6600",
          "axisThickness": 2,
          "axisAlpha": 1,
          "position": "left"
        },/* {
          "id":"v2",
          "axisColor": "#FCD202",
          "axisThickness": 2,
          "axisAlpha": 1,
          "position": "right"
        },*/ {
          "id":"v3",
          "axisColor": "#B0DE09",
          "axisThickness": 2,
          "gridAlpha": 0,
          "offset": 50,
          "axisAlpha": 1,
          "position": "left"
        }],
        "graphs": [{
          "valueAxis": "v1",
          "lineColor": "#FF6600",
          "bullet": "round",
          "bulletBorderThickness": 1,
          "hideBulletsCount": 30,
          "title": "失败次数",
          "valueField": "fail",
          "fillAlphas": 0,
          "balloonText": "[[day]]<br><b><span style='font-size:14px;'>失败次数:[[fail]]</span></b>",
        },/* {
          "valueAxis": "v2",
          "lineColor": "#FCD202",
          "bullet": "square",
          "bulletBorderThickness": 1,
          "hideBulletsCount": 30,
          "title": "yellow line",
          "valueField": "hits",
          "fillAlphas": 0
        }, */{
          "valueAxis": "v3",
          "lineColor": "#B0DE09",
          "bullet": "triangleUp",
          "bulletBorderThickness": 1,
          "hideBulletsCount": 30,
          "title": "执行次数",
          "valueField": "success",
          "fillAlphas": 0,
          "balloonText": "[[day]]<br><b><span style='font-size:14px;'>执行次数:[[success]]</span></b>",
        }],
        "chartScrollbar": {},
        "chartCursor": {
          "cursorPosition": "mouse"
        },
        "categoryField": "day",
        "categoryAxis": {
          "parseDates": false,
          "axisColor": "#DADADA",
          "minorGridEnabled": true
        },
        "export": {
          "enabled": true,
          "position": "bottom-right"
        }
      });

      function zoomChart(){
        chart.zoomToIndexes(0, chart.dataProvider.length - 1);
      }
      chart.addListener("dataUpdated", zoomChart);
      zoomChart();

// generate some random data, quite different range
//       function generateChartData() {
//         var chartData = [];
//         var firstDate = new Date();
//         firstDate.setDate(firstDate.getDate() - 100);
//
//         var visits = 1600;
//         var hits = 2900;
//         var views = 8700;
//
//
//         for (var i = 0; i < 100; i++) {
//           // we create date objects here. In your data, you can have date strings
//           // and then set format of your dates using chart.dataDateFormat property,
//           // however when possible, use date objects, as this will speed up chart rendering.
//           var newDate = new Date(firstDate);
//           newDate.setDate(newDate.getDate() + i);
//
//           visits += Math.round((Math.random()<0.5?1:-1)*Math.random()*10);
//           views += Math.round((Math.random()<0.5?1:-1)*Math.random()*10);
//
//           chartData.push({
//             date: newDate,
//             visits: visits,
//             views: views
//           });
//         }
//         return chartData;
//       }
//
//       function zoomChart(){
//         chart.zoomToIndexes(0, chart.dataProvider.length - 1);
//       }


      that.getCharts(chart)
      that.getStatistics()
      window.setInterval(function () {
        that.getStatistics()
        that.getCharts(chart)
      }, 2000)
    },
    created: function () {
      // 视图渲染后，追加需要的js、css
      var script = document.createElement("script");
      script.src = "./static/js/moment-2.2.1.js";
      document.body.appendChild(script);

      script = document.createElement("script");
      script.src = "./static/js/protovis-d3.2.js";
      document.body.appendChild(script);
      script = document.createElement("script");
      script.src = "./static/js/Chart.js";
      document.body.appendChild(script);

      script = document.createElement("script");
      script.src = "./static/js/vix.js";
      document.body.appendChild(script);


      script = document.createElement("script");
      script.src = "./static/js/index.vue.js";
      document.body.appendChild(script)

      var link = document.createElement("link")
      link.rel = "stylesheet"
      link.src = "./static/css/vroom.css"
      document.body.appendChild(link)

    },
  }
</script>
<style>
  #chartdiv {
    width	: 100%;
  }
</style>
