<template>
  <!--custom-widgets-->
  <div>
    <div class="custom-widgets" author="yuyi">
      <div class="row-one">
        <div class="col-md-3 widget">
          <div class="stats-left ">
            <h5>Today</h5>
            <h4> Users</h4>
          </div>
          <div class="stats-right">
            <label>90</label>
          </div>
          <div class="clearfix"> </div>
        </div>
        <div class="col-md-3 widget states-mdl">
          <div class="stats-left">
            <h5>Today</h5>
            <h4>Visitors</h4>
          </div>
          <div class="stats-right">
            <label> 85</label>
          </div>
          <div class="clearfix"> </div>
        </div>
        <div class="col-md-3 widget states-thrd">
          <div class="stats-left">
            <h5>Today</h5>
            <h4>Tasks</h4>
          </div>
          <div class="stats-right">
            <label>51</label>
          </div>
          <div class="clearfix"> </div>
        </div>
        <div class="col-md-3 widget states-last">
          <div class="stats-left">
            <h5>Today</h5>
            <h4>Alerts</h4>
          </div>
          <div class="stats-right">
            <label>30</label>
          </div>
          <div class="clearfix"> </div>
        </div>
        <div class="clearfix"> </div>
      </div>
    </div>
    <!--//custom-widgets-->
    <!--/candile-->
    <div class="candile">
      <div class="candile-inner">
        <h3 class="sub-tittle">Candlestick Chart </h3>
        <div id="center"><div id="fig">
          <script type="text/javascript+protovis">

															/* Parse dates. */
															var dateFormat = pv.Format.date("%d-%b-%y");
															vix.forEach(function(d) d.date = dateFormat.parse(d.date));

															/* Scales. */
															var w =1220,
																h = 300,
																x = pv.Scale.linear(vix, function(d) d.date).range(0, w),
																y = pv.Scale.linear(vix, function(d) d.low, function(d) d.high).range(0, h).nice();

															var vis = new pv.Panel()
																.width(w)
																.height(h)
																.margin(10)
																.left(30);

															/* Dates. */
															vis.add(pv.Rule)
																 .data(x.ticks())
																 .left(x)
																 .strokeStyle("#eee")
															   .anchor("bottom").add(pv.Label)
																 .text(x.tickFormat);

															/* Prices. */
															vis.add(pv.Rule)
																 .data(y.ticks(7))
																 .bottom(y)
																 .left(-10)
																 .right(-10)
																 .strokeStyle(function(d) d % 10 ? "#ddd" : "#ddd")
															   .anchor("left").add(pv.Label)
																 .textStyle(function(d) d % 10 ? "#999" : "#ddd")
																 .text(y.tickFormat);

															/* Candlestick. */
															vis.add(pv.Rule)
																.data(vix)
																.left(function(d) x(d.date))
																.bottom(function(d) y(Math.min(d.high, d.low)))
																.height(function(d) Math.abs(y(d.high) - y(d.low)))
																.strokeStyle(function(d) d.open < d.close ? "#052963" : "#00C6D7")
															  .add(pv.Rule)
																.bottom(function(d) y(Math.min(d.open, d.close)))
																.height(function(d) Math.abs(y(d.open) - y(d.close)))
																.lineWidth(10);

															vis.render();

																</script>

        </div>
        </div>

      </div>

    </div>
    <!--/candile-->

    <!--/charts-->
    <div class="charts">
      <div class="chrt-inner">
        <div class="chrt-bars">
          <div class="col-md-6 chrt-two">
            <h3 class="sub-tittle">Bar Chart </h3>
            <div id="chart2"></div>

          </div>
          <div class="col-md-6 chrt-three">
            <h3 class="sub-tittle">Error Chart </h3>
            <div id="chartdiv"></div>
            <script>
            </script>


          </div>
          <div class="clearfix"> </div>
        </div>
        <!--/float-charts-->
        <div class="pie">
          <div class="col-md-6 chrt-two">
            <h3 class="sub-tittle">Reversed Value Axis Chart</h3>
            <div id="chartdiv2"></div>
            <script>
            </script>

          </div>
          <div class="col-md-6 chrt-three second">
            <h3 class="sub-tittle">Radar chart</h3>
            <div id="chartdiv4"></div>
            <script>
            </script>

          </div>
          <div class="clearfix"> </div>
        </div>
        <div class="area">
          <div class="col-md-6 chrt-two area">
            <h3 class="sub-tittle">Line Multi Chart</h3>
            <div style="area">
              <canvas id="canvas" style="width:100%;height:100%"></canvas>
            </div>
            <button id="randomizeData">Randomize Data</button>


          </div>
          <div class="col-md-6 chrt-three">
            <h3 class="sub-tittle">Stacked bar chart</h3>
            <div id="chartdiv1"></div>
          </div>
          <div class="clearfix"></div>
        </div>
        <!--//weather-charts-->
        <div class="graph-visualization">
          <div class="col-md-6 map-1">
            <h3 class="sub-tittle">Weather </h3>
            <div class="weather">
              <div class="weather-top">
                <div class="weather-top-left">
                  <div class="degree">
                    <figure class="icons">
                      <canvas id="partly-cloudy-day" width="64" height="64">
                      </canvas>
                    </figure>
                    <span>37°</span>
                    <div class="clearfix"></div>
                  </div>
                  <script>

                  </script>
                  <p>FRIDAY
                    <label>13</label>
                    <sup>th</sup>
                    AUG
                  </p>
                </div>
                <div class="weather-top-right">
                  <p><i class="fa fa-map-marker"></i>Location</p>
                  <label>Lorem ipsum</label>
                </div>
                <div class="clearfix"> </div>
              </div>
              <div class="weather-bottom">
                <div class="weather-bottom1">
                  <div class="weather-head">
                    <h4>Cloudy</h4>
                    <figure class="icons">
                      <canvas id="cloudy" width="40" height="40"></canvas>
                    </figure>
                    <script>

                    </script>
                    <h6>20°</h6>
                    <div class="bottom-head">
                      <p>Monday</p>
                    </div>
                  </div>
                </div>
                <div class="weather-bottom1 ">
                  <div class="weather-head">
                    <h4>Sunny</h4>
                    <figure class="icons">
                      <canvas id="clear-day" width="40" height="40">
                      </canvas>

                    </figure>
                    <script>

                    </script>
                    <h6>37°</h6>
                    <div class="bottom-head">
                      <p>Tuesday</p>
                    </div>
                  </div>
                </div>
                <div class="weather-bottom1">
                  <div class="weather-head">
                    <h4>Rainy</h4>
                    <figure class="icons">
                      <canvas id="sleet" width="40" height="40">
                      </canvas>
                    </figure>
                    <script>

                    </script>

                    <h6>7°</h6>
                    <div class="bottom-head">
                      <p>Wednesday</p>
                    </div>
                  </div>
                </div>
                <div class="weather-bottom1 ">
                  <div class="weather-head">
                    <h4>Snowy</h4>
                    <figure class="icons">
                      <canvas id="snow" width="40" height="40">
                      </canvas>
                    </figure>
                    <script>

                    </script>
                    <h6>-10°</h6>
                    <div class="bottom-head">
                      <p>Thursday</p>
                    </div>
                  </div>
                </div>
                <div class="clearfix"> </div>
              </div>

            </div>

          </div>
          <div class="col-md-6 map-2">
            <div class="profile-nav alt">
              <section class="panel">
                <div class="user-heading alt clock-row terques-bg">
                  <h3 class="sub-tittle clock">Easy Clock </h3>
                </div>
                <ul id="clock">
                  <li id="sec"></li>
                  <li id="hour"></li>
                  <li id="min"></li>
                </ul>

                <ul class="clock-category">
                  <li>
                    <a href="#" class="active">
                      <img src="../../../static/images/time.png" alt="">
                      <span>Clock</span>
                    </a>
                  </li>
                  <li>
                    <a href="#">
                      <img src="../../../static/images/alarm.png" alt="">
                      <span>Alarm</span>
                    </a>
                  </li>
                  <li>
                    <a href="#">
                      <img src="../../../static/images/watch.png" alt="">
                      <span>Stop watch</span>
                    </a>
                  </li>
                  <li>
                    <a href="#">
                      <img src="../../../static/images/timer.png" alt="">
                      <span>Timer</span>
                    </a>
                  </li>
                </ul>

              </section>

            </div>
          </div>
          <div class="clearfix"> </div>
        </div>

        <!--//charts-->
        <div class="area-charts">
          <div class="col-md-6 panel-chrt">
            <h3 class="sub-tittle dyna">Dynamic Speedometer </h3>
            <div id="wrapper">
              <div id="left">
                <div>
                  <label for='miles'>Miles:</label> <input type="text" name="miles" id="miles" />
                </div>
                <div>
                  <label for='kilometers'>Kilometers:</label> <input type="text" name="kilometers" id="kilometers" />
                </div>
                <p id="errmsg"></p>
              </div>

              <div id="gauge">
                <div id="circle"></div>
                <div id="numbers"></div>
                <div id="mi-km"></div>
                <div id="needle"></div>
              </div>
            </div>
          </div>
          <div class="col-md-6 tini-time-line">
            <h3 class="sub-tittle">Time line </h3>
            <ul class="timeline">
              <li>
                <div class="timeline-badge info"><i class="fa fa-smile-o"></i></div>
                <div class="timeline-panel">
                  <div class="timeline-heading">
                    <h4 class="timeline-title">Title</h4>
                  </div>
                  <div class="timeline-body">
                    <p>Description...</p>
                  </div>
                </div>
              </li>

              <li>
                <div class="timeline-badge primary"><i class="fa fa-star-o"></i></div>
                <div class="timeline-panel">
                  <div class="timeline-heading">
                    <h4 class="timeline-title">Title</h4>
                  </div>
                  <div class="timeline-body">
                    <p>Description...</p>
                  </div>
                </div>
              </li>



              <li>
                <div class="timeline-badge danger"><i class="fa fa-times-circle-o"></i></div>
                <div class="timeline-panel">
                  <div class="timeline-heading">
                    <h4 class="timeline-title">Title</h4>
                  </div>
                  <div class="timeline-body">
                    <p>Description...</p>
                  </div>
                </div>
              </li>

              <li>
                <div class="timeline-badge success"><i class="fa fa-check-circle-o"></i></div>
                <div class="timeline-panel">
                  <div class="timeline-heading">
                    <h4 class="timeline-title">Title</h4>
                  </div>
                  <div class="timeline-body">
                    <p>Description...</p>
                  </div>
                </div>
              </li>
            </ul>
          </div>
          <div class="clearfix"></div>
        </div>
        <!--/bottom-grids-->
        <div class="bottom-grids">
          <div class="dev-table">
            <div class="col-md-4 dev-col">

              <div class="dev-widget dev-widget-transparent">
                <h2 class="inner one">Server Usage</h2>
                <p>Today server usage in percentages</p>
                <div class="dev-stat"><span class="counter">68</span>%</div>
                <div class="progress progress-bar-xs">
                  <div class="progress-bar progress-bar-info" role="progressbar" aria-valuenow="60" aria-valuemin="0" aria-valuemax="100" style="width: 50%;"></div>
                </div>
                <p>We Todayly recommend you change your plan to <strong>Pro</strong>. Click here to get more details.</p>

                <a href="#" class="dev-drop">Take a closer look <span class="fa fa-angle-right pull-right"></span></a>
              </div>

            </div>
            <div class="col-md-4 dev-col mid">

              <div class="dev-widget dev-widget-transparent dev-widget-success">
                <h3 class="inner">Today Earnings</h3>
                <p>This is Today earnings per last year</p>
                <div class="dev-stat">$<span class="counter">75,332</span></div>
                <div class="progress progress-bar-xs">
                  <div class="progress-bar progress-bar-success" role="progressbar" aria-valuenow="60" aria-valuemin="0" aria-valuemax="100" style="width: 79%;"></div>
                </div>
                <p>We happy to notice you that you become an Elite customer, and you can get some custom sugar.</p>

                <a href="#" class="dev-drop">Take a closer look <span class="fa fa-angle-right pull-right"></span></a>
              </div>

            </div>
            <div class="col-md-4 dev-col">

              <div class="dev-widget dev-widget-transparent dev-widget-danger">
                <h3 class="inner">Your Balance</h3>
                <p>All your earnings for this time</p>
                <div class="dev-stat">$<span class="counter">5,321</span></div>
                <div class="progress progress-bar-xs">
                  <div class="progress-bar progress-bar-danger" role="progressbar" aria-valuenow="60" aria-valuemin="0" aria-valuemax="100" style="width: 80%;"></div>
                </div>
                <p>You can withdraw this money in end of each month. Also you can spend it on our marketplace.</p>

                <a href="#" class="dev-drop">Take a closer look <span class="fa fa-angle-right pull-right"></span></a>
              </div>

            </div>
            <div class="clearfix"></div>

          </div>
        </div>
        <!--//bottom-grids-->

      </div>
      <!--/charts-inner-->
    </div>
  </div>
  <!--//outer-wp-->
</template>
<script>
  export default {
    name: "Index",
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
