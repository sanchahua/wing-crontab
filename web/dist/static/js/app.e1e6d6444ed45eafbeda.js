webpackJsonp([0],{A5SV:function(t,e){},"Ces+":function(t,e,s){"use strict";(function(t){e.a={name:"Edit",data:function(){return{cron_info:{cron_set:""}}},mounted:function(){var t=this;jeDate("#start-time",{festival:!0,minDate:"1900-01-01",maxDate:"2099-12-31",method:{choose:function(t){}},format:"YYYY-MM-DD hh:mm:ss",toggle:function(t){},donefun:function(e){console.log(e),t.cron_info.start_time=e.val}}),jeDate("#end-time",{festival:!0,minDate:"1900-01-01",maxDate:"2099-12-31",method:{choose:function(t){}},format:"YYYY-MM-DD hh:mm:ss",toggle:function(t){},donefun:function(e){console.log(e),t.cron_info.end_time=e.val}}),this.getInfo()},methods:{getInfo:function(){var t=window.location.hash.split("?",-1);console.log(t);var e=this,s={};if(t.length>1){var a=t[1].split("&"),i=0,o=a.length;for(i=0;i<o;i++){var n=a[i].split("=");n.length>1&&(s[n[0]]=n[1])}console.log(s)}void 0!==s.id&&axios.get("/cron/info/"+s.id).then(function(t){2e3==t.data.code?(console.log(t),e.cron_info=t.data.data):alert(t.data.message)}).catch(function(t){})},submit:function(){var e={cron_set:this.cron_info.cron_set,command:this.cron_info.command,start_time:t("#start-time").val(),end_time:t("#end-time").val(),remark:this.cron_info.remark,stop:this.cron_info.stop?1:0,is_mutex:this.cron_info.is_mutex?1:0};console.log(this.cron_info,e);var s=window.location.hash.split("?",-1);console.log(s);var a={};if(s.length>1){var i=s[1].split("&"),o=0,n=i.length;for(o=0;o<n;o++){var r=i[o].split("=");r.length>1&&(a[r[0]]=r[1])}console.log(a)}axios.post("/cron/update/"+a.id,e).then(function(t){console.log(t),2e3==t.data.code?window.location.href="/ui/#/cron_list":alert(t.data.message)}).catch(function(t){alert(t)})}},created:function(){}}}).call(e,s("ZosL"))},DnYX:function(t,e,s){"use strict";(function(t){e.a={name:"CronList",data:function(){return{cron_list:[],is_stop:0,is_mutex:0,is_timeout:0,keyword:""}},mounted:function(){this.list()},methods:{list:function(){var t=this,e="/cron/list?";0==t.is_stop?e+="stop=&":1==t.is_stop?e+="stop=1&":2==t.is_stop&&(e+="stop=0&"),0==t.is_mutex?e+="mutex=&":1==t.is_mutex?e+="mutex=1&":2==t.is_mutex&&(e+="mutex=0&"),0==t.is_timeout?e+="timeout=&":1==t.is_timeout?e+="timeout=1&":2==t.is_timeout&&(e+="timeout=0&"),e+="keyword="+encodeURIComponent(t.keyword),console.log(e),axios.get(e).then(function(e){2e3==e.data.code?(console.log(e.data.data),t.cron_list=e.data.data):alert(e.data.message)}).catch(function(t){})},stop:function(e){var s=t(e.target).attr("item-id"),a=this;axios.get("/cron/stop/"+s).then(function(t){if(2e3==t.data.code){console.log(t);var e=a.cron_list.length,i=0;for(i=0;i<e;i++)a.cron_list[i].id==s&&(a.cron_list[i].stop=!0,console.log(a.cron_list[i]))}else alert(t.data.message)}).catch(function(t){})},start:function(e){var s=t(e.target).attr("item-id"),a=this;axios.get("/cron/start/"+s).then(function(t){if(2e3==t.data.code){console.log(t);var e=a.cron_list.length,i=0;for(i=0;i<e;i++)a.cron_list[i].id==s&&(a.cron_list[i].stop=!1,console.log(a.cron_list[i]))}else alert(t.data.message)}).catch(function(t){})},del:function(e){var s=t(e.target).attr("item-id");if(window.confirm("确认删除"+s+"？")){var a=this;axios.get("/cron/delete/"+s).then(function(t){if(2e3==t.data.code){console.log(t);var e=a.cron_list.length,i=0;for(i=0;i<e;i++)a.cron_list[i].id==s&&(console.log(a.cron_list[i]),a.cron_list.splice(i,1))}else alert(t.data.message)}).catch(function(t){})}},jumpEdit:function(e){var s=t(e.target).attr("item-id");window.location.href="/ui/#/edit?id="+s},jumpLogs:function(){var e=t(event.target).attr("item-id");window.location.href="/ui/#/logs?id="+e},mutex0:function(e){var s=t(e.target).attr("item-id"),a=this;axios.get("/cron/mutex/false/"+s).then(function(t){if(2e3==t.data.code){console.log(t);var e=a.cron_list.length,i=0;for(i=0;i<e;i++)a.cron_list[i].id==s&&(a.cron_list[i].is_mutex=!1,console.log(a.cron_list[i]))}else alert(t.data.message)}).catch(function(t){})},mutex1:function(e){var s=t(e.target).attr("item-id"),a=this;axios.get("/cron/mutex/true/"+s).then(function(t){if(2e3==t.data.code){console.log(t);var e=a.cron_list.length,i=0;for(i=0;i<e;i++)a.cron_list[i].id==s&&(a.cron_list[i].is_mutex=!0,console.log(a.cron_list[i]))}else alert(t.data.message)}).catch(function(t){})},analysis:function(){},run:function(e){var s=t(e.target).attr("item-id"),a=t(e.target).attr("item-command");dialog({title:"确认执行？",content:'<div style="color: #f00;">注意：即将执行指令['+a+']，最终产生的风险(系统风险、业务风险)请自行评估！</div><div>默认超时时间：<input id="property-returnValue-demo" value="3" />秒</div><div class="run-command-response">执行返回值：</div><div class="run-command-response"><textarea style="width: 100%;" id="run-command-response-show"></textarea></div>',ok:function(){var e=this;e.title("执行中…");var a=t("#property-returnValue-demo").val();return axios.get("/cron/run/"+s+"/"+a).then(function(s){2e3==s.data.code?(console.log(s),t("#run-command-response-show").val(s.data.data)):t("#run-command-response-show").val(s.data.message),e.title("执行完成")}).catch(function(e){t("#run-command-response-show").val("网络异常")}),!1},okValue:"确定",cancelValue:"取消",cancel:function(){}}).show()},quick_search:function(t){var e=this;window.setTimeout(function(){console.log("stop=",e.is_stop),console.log("is_mutex=",e.is_mutex),console.log("is_timeout=",e.is_timeout),console.log("keyword=",e.keyword),e.list()},500),t.stopPropagation()},search_keyword:function(){var t=this;window.setTimeout(function(){t.list()},500)}},created:function(){}}}).call(e,s("ZosL"))},LneM:function(t,e){},NHnr:function(t,e,s){"use strict";Object.defineProperty(e,"__esModule",{value:!0});var a=s("cAeh"),i={render:function(){var t=this.$createElement,e=this._self._c||t;return e("div",{staticClass:"outter-wp"},[e("router-view")],1)},staticRenderFns:[]},o=s("Mz/3")({name:"App"},i,!1,null,null,null).exports,n=s("1eSk"),r={render:function(){var t=this,e=t.$createElement,s=t._self._c||e;return s("div",{staticClass:"hello"},[s("h1",[t._v(t._s(t.msg))]),t._v(" "),s("h2",[t._v("Essential Links")]),t._v(" "),t._m(0),t._v(" "),s("h2",[t._v("Ecosystem")]),t._v(" "),t._m(1)])},staticRenderFns:[function(){var t=this,e=t.$createElement,s=t._self._c||e;return s("ul",[s("li",[s("a",{attrs:{href:"https://vuejs.org",target:"_blank"}},[t._v("\n        Core Docs\n      ")])]),t._v(" "),s("li",[s("a",{attrs:{href:"https://forum.vuejs.org",target:"_blank"}},[t._v("\n        Forum\n      ")])]),t._v(" "),s("li",[s("a",{attrs:{href:"https://chat.vuejs.org",target:"_blank"}},[t._v("\n        Community Chat\n      ")])]),t._v(" "),s("li",[s("a",{attrs:{href:"https://twitter.com/vuejs",target:"_blank"}},[t._v("\n        Twitter\n      ")])]),t._v(" "),s("br"),t._v(" "),s("li",[s("a",{attrs:{href:"http://vuejs-templates.github.io/webpack/",target:"_blank"}},[t._v("\n        Docs for This Template\n      ")])])])},function(){var t=this.$createElement,e=this._self._c||t;return e("ul",[e("li",[e("a",{attrs:{href:"http://router.vuejs.org/",target:"_blank"}},[this._v("\n        vue-router\n      ")])]),this._v(" "),e("li",[e("a",{attrs:{href:"http://vuex.vuejs.org/",target:"_blank"}},[this._v("\n        vuex\n      ")])]),this._v(" "),e("li",[e("a",{attrs:{href:"http://vue-loader.vuejs.org/",target:"_blank"}},[this._v("\n        vue-loader\n      ")])]),this._v(" "),e("li",[e("a",{attrs:{href:"https://github.com/vuejs/awesome-vue",target:"_blank"}},[this._v("\n        awesome-vue\n      ")])])])}]};s("Mz/3")({name:"HelloWorld",data:function(){return{msg:"Welcome to Your Vue.js App"}}},r,!1,function(t){s("LneM")},"data-v-d8ec41bc",null).exports;var c={name:"Index",data:function(){return{days:7,cron_count:0,history_run_count:0,day_run_count:0,day_run_fail_count:0}},methods:{getStatistics:function(){var t=this;axios.get("/index").then(function(e){2e3==e.data.code?(console.log(e),t.cron_count=e.data.data.cron_count,t.history_run_count=e.data.data.history_run_count,t.day_run_count=e.data.data.day_run_count,t.day_run_fail_count=e.data.data.day_run_fail_count):alert(e.data.message)}).catch(function(t){})},getCharts:function(t){axios.get("/charts/"+this.days).then(function(e){console.log(e),2e3==e.data.code?(t.dataProvider=e.data.data,t.validateNow(),t.validateData()):alert(e.data.message)}).catch(function(t){})}},mounted:function(){var t=this,e=AmCharts.makeChart("chartdiv",{type:"serial",theme:"light",dataDateFormat:"YYYY-MM-DD",legend:{useGraphSettings:!0},dataProvider:[],synchronizeGrid:!0,valueAxes:[{id:"v1",axisColor:"#FF6600",axisThickness:2,axisAlpha:1,position:"left"},{id:"v3",axisColor:"#B0DE09",axisThickness:2,gridAlpha:0,offset:50,axisAlpha:1,position:"left"}],graphs:[{valueAxis:"v1",lineColor:"#FF6600",bullet:"round",bulletBorderThickness:1,hideBulletsCount:30,title:"失败次数",valueField:"fail",fillAlphas:0,balloonText:"[[day]]<br><b><span style='font-size:14px;'>失败次数:[[fail]]</span></b>"},{valueAxis:"v3",lineColor:"#B0DE09",bullet:"triangleUp",bulletBorderThickness:1,hideBulletsCount:30,title:"执行次数",valueField:"success",fillAlphas:0,balloonText:"[[day]]<br><b><span style='font-size:14px;'>执行次数:[[success]]</span></b>"}],chartScrollbar:{},chartCursor:{cursorPosition:"mouse"},categoryField:"day",categoryAxis:{parseDates:!1,axisColor:"#DADADA",minorGridEnabled:!0},export:{enabled:!0,position:"bottom-right"}});function s(){e.zoomToIndexes(0,e.dataProvider.length-1)}e.addListener("dataUpdated",s),s(),t.getCharts(e),t.getStatistics(),window.setInterval(function(){t.getStatistics(),t.getCharts(e)},2e3)},created:function(){var t=document.createElement("script");t.src="./static/js/moment-2.2.1.js",document.body.appendChild(t),(t=document.createElement("script")).src="./static/js/protovis-d3.2.js",document.body.appendChild(t),(t=document.createElement("script")).src="./static/js/Chart.js",document.body.appendChild(t),(t=document.createElement("script")).src="./static/js/vix.js",document.body.appendChild(t),(t=document.createElement("script")).src="./static/js/index.vue.js",document.body.appendChild(t);var e=document.createElement("link");e.rel="stylesheet",e.src="./static/css/vroom.css",document.body.appendChild(e)}},l={render:function(){var t=this,e=t.$createElement,s=t._self._c||e;return s("div",[s("div",{staticClass:"custom-widgets",attrs:{author:"yuyi"}},[s("div",{staticClass:"row-one"},[s("div",{staticClass:"col-md-3 widget"},[t._m(0),t._v(" "),s("div",{staticClass:"stats-right"},[s("label",[t._v(t._s(t.cron_count))])]),t._v(" "),s("div",{staticClass:"clearfix"})]),t._v(" "),s("div",{staticClass:"col-md-3 widget states-mdl"},[t._m(1),t._v(" "),s("div",{staticClass:"stats-right"},[s("label",[t._v(t._s(t.history_run_count))])]),t._v(" "),s("div",{staticClass:"clearfix"})]),t._v(" "),s("div",{staticClass:"col-md-3 widget states-thrd"},[t._m(2),t._v(" "),s("div",{staticClass:"stats-right"},[s("label",[t._v(t._s(t.day_run_count))])]),t._v(" "),s("div",{staticClass:"clearfix"})]),t._v(" "),s("div",{staticClass:"col-md-3 widget states-last"},[t._m(3),t._v(" "),s("div",{staticClass:"stats-right"},[s("label",[t._v(t._s(t.day_run_fail_count))])]),t._v(" "),s("div",{staticClass:"clearfix"})]),t._v(" "),s("div",{staticClass:"clearfix"})])]),t._v(" "),s("div",{staticClass:"charts"},[s("div",{staticClass:"chrt-inner"},[s("div",{staticClass:"chrt-bars"},[s("div",{staticClass:"candile-inner"},[s("h3",{staticClass:"sub-tittle sub-tittle-e"},[t._v("执行次数 ")]),t._v(" "),s("span",{staticClass:"time-tool"},[s("label",[s("input",{directives:[{name:"model",rawName:"v-model",value:t.days,expression:"days"}],attrs:{name:"time-pass",value:"7",type:"radio",checked:""},domProps:{checked:t._q(t.days,"7")},on:{change:function(e){t.days="7"}}}),t._v("过去一周")]),t._v(" "),s("label",[s("input",{directives:[{name:"model",rawName:"v-model",value:t.days,expression:"days"}],attrs:{name:"time-pass",value:"30",type:"radio"},domProps:{checked:t._q(t.days,"30")},on:{change:function(e){t.days="30"}}}),t._v("过去一个月")]),t._v(" "),s("label",[s("input",{directives:[{name:"model",rawName:"v-model",value:t.days,expression:"days"}],attrs:{name:"time-pass",value:"90",type:"radio"},domProps:{checked:t._q(t.days,"90")},on:{change:function(e){t.days="90"}}}),t._v("过去三个月")]),t._v(" "),s("label",[s("input",{directives:[{name:"model",rawName:"v-model",value:t.days,expression:"days"}],attrs:{name:"time-pass",value:"180",type:"radio"},domProps:{checked:t._q(t.days,"180")},on:{change:function(e){t.days="180"}}}),t._v("过去六个月")]),t._v(" "),s("label",[s("input",{directives:[{name:"model",rawName:"v-model",value:t.days,expression:"days"}],attrs:{name:"time-pass",value:"365",type:"radio"},domProps:{checked:t._q(t.days,"365")},on:{change:function(e){t.days="365"}}}),t._v("过去一年")])]),t._v(" "),s("div",{attrs:{id:"chartdiv"}})]),t._v(" "),s("div",{staticClass:"clearfix"})])])])])},staticRenderFns:[function(){var t=this.$createElement,e=this._self._c||t;return e("div",{staticClass:"stats-left "},[e("h5",[this._v("定时任务")]),this._v(" "),e("h4",[this._v(" 总数")])])},function(){var t=this.$createElement,e=this._self._c||t;return e("div",{staticClass:"stats-left"},[e("h5",[this._v("历史执行")]),this._v(" "),e("h4",[this._v("次数")])])},function(){var t=this.$createElement,e=this._self._c||t;return e("div",{staticClass:"stats-left"},[e("h5",[this._v("今日执行")]),this._v(" "),e("h4",[this._v("次数")])])},function(){var t=this.$createElement,e=this._self._c||t;return e("div",{staticClass:"stats-left"},[e("h5",[this._v("今日错误")]),this._v(" "),e("h4",[this._v("次数")])])}]};var d=s("Mz/3")(c,l,!1,function(t){s("e1hH")},null,null).exports,_={name:"Add",data:function(){return{datetime:(new Date).Format("yyyy-MM-dd hh:mm:ss")}},created:function(){var t=document.createElement("script");t.src="./static/js/add.vue.js?t="+(new Date).valueOf(),document.body.appendChild(t)}},v={render:function(){var t=this,e=t.$createElement,s=t._self._c||e;return s("div",[t._m(0),t._v(" "),s("div",{staticClass:"forms-main"},[s("h2",{staticClass:"inner-tittle"},[t._v("增加定时任务 ")]),t._v(" "),s("div",{staticClass:"graph-form"},[s("div",{staticClass:"form-body"},[t._m(1),t._v(" "),s("div",{staticClass:"form-group"},[s("label",{attrs:{for:"start-time"}},[t._v("开始时间，大于等于此时间才执行，不限留空")]),t._v(" "),s("input",{staticClass:"form-control",attrs:{type:"text",id:"start-time"},domProps:{value:t.datetime}})]),t._v(" "),t._m(2),t._v(" "),t._m(3),t._v(" "),t._m(4),t._v(" "),t._m(5),t._v(" "),t._m(6),t._v(" "),s("button",{staticClass:"btn btn-default",attrs:{type:"button",id:"do-submit"}},[t._v("提交")])])])])])},staticRenderFns:[function(){var t=this.$createElement,e=this._self._c||t;return e("div",{staticClass:"sub-heard-part"},[e("ol",{staticClass:"breadcrumb m-b-0"},[e("li",[e("a",{attrs:{href:"/ui/#/"}},[this._v("首页")])]),this._v(" "),e("li",{staticClass:"active"},[this._v("增加定时任务")])])])},function(){var t=this.$createElement,e=this._self._c||t;return e("div",{staticClass:"form-group"},[e("label",{attrs:{for:"cron-set"}},[this._v("定时配置，如：*/1 * * * * *，这里精确到秒，前面的意思是每秒执行一次，分别对应，秒分时日月周")]),this._v(" "),e("input",{staticClass:"form-control",attrs:{type:"text",id:"cron-set"}})])},function(){var t=this.$createElement,e=this._self._c||t;return e("div",{staticClass:"form-group"},[e("label",{attrs:{for:"end-time"}},[this._v("结束时间，小于此时间才执行，不限留空")]),this._v(" "),e("input",{staticClass:"form-control",attrs:{type:"text",id:"end-time",value:"2099-01-01 08:00:00"}})])},function(){var t=this.$createElement,e=this._self._c||t;return e("div",{staticClass:"form-group"},[e("label",{attrs:{for:"command"}},[this._v("执行命令")]),this._v(" "),e("input",{staticClass:"form-control",attrs:{type:"text",id:"command",value:""}})])},function(){var t=this.$createElement,e=this._self._c||t;return e("div",{staticClass:"checkbox"},[e("label",[e("input",{attrs:{type:"checkbox",id:"cron-stop"}}),this._v(" 初始化为停止状态\n          ")])])},function(){var t=this.$createElement,e=this._self._c||t;return e("div",{staticClass:"checkbox"},[e("label",[e("input",{attrs:{type:"checkbox",id:"cron-is-mutex"}}),this._v(" 严格互斥执行\n          ")])])},function(){var t=this.$createElement,e=this._self._c||t;return e("div",{staticClass:"form-group"},[e("label",{attrs:{for:"remark"}},[this._v("备注")]),this._v(" "),e("textarea",{staticClass:"form-control",attrs:{id:"remark"}})])}]},u=s("Mz/3")(_,v,!1,null,null,null).exports,m=s("DnYX"),h={render:function(){var t=this,e=t.$createElement,s=t._self._c||e;return s("div",[t._m(0),t._v(" "),s("div",{staticClass:"graph-visual tables-main"},[s("h3",{staticClass:"inner-tittle two"},[t._v("定时任务列表（"+t._s(t.cron_list.length)+"个） ")]),t._v(" "),s("div",{staticClass:"search-tool"},[s("div",[s("input",{directives:[{name:"model",rawName:"v-model",value:t.keyword,expression:"keyword"}],attrs:{type:"text"},domProps:{value:t.keyword},on:{input:function(e){e.target.composing||(t.keyword=e.target.value)}}}),s("input",{attrs:{type:"button",value:"查询"},on:{click:t.search_keyword}})]),t._v(" "),s("div",{on:{click:t.quick_search}},[s("div",[s("label",[s("input",{directives:[{name:"model",rawName:"v-model",value:t.is_stop,expression:"is_stop"}],attrs:{name:"is-stop",value:"2",type:"radio"},domProps:{checked:t._q(t.is_stop,"2")},on:{change:function(e){t.is_stop="2"}}}),t._v("正在运行")]),t._v(" "),s("label",[s("input",{directives:[{name:"model",rawName:"v-model",value:t.is_stop,expression:"is_stop"}],attrs:{name:"is-stop",value:"1",type:"radio"},domProps:{checked:t._q(t.is_stop,"1")},on:{change:function(e){t.is_stop="1"}}}),t._v("已停止")]),t._v(" "),s("label",[s("input",{directives:[{name:"model",rawName:"v-model",value:t.is_stop,expression:"is_stop"}],attrs:{name:"is-stop",value:"0",type:"radio"},domProps:{checked:t._q(t.is_stop,"0")},on:{change:function(e){t.is_stop="0"}}}),t._v("取消")])]),t._v(" "),s("div",[s("label",[s("input",{directives:[{name:"model",rawName:"v-model",value:t.is_mutex,expression:"is_mutex"}],attrs:{name:"is-mutex",value:"1",type:"radio"},domProps:{checked:t._q(t.is_mutex,"1")},on:{change:function(e){t.is_mutex="1"}}}),t._v("互斥")]),t._v(" "),s("label",[s("input",{directives:[{name:"model",rawName:"v-model",value:t.is_mutex,expression:"is_mutex"}],attrs:{name:"is-mutex",value:"2",type:"radio"},domProps:{checked:t._q(t.is_mutex,"2")},on:{change:function(e){t.is_mutex="2"}}}),t._v("非互斥")]),t._v(" "),s("label",[s("input",{directives:[{name:"model",rawName:"v-model",value:t.is_mutex,expression:"is_mutex"}],attrs:{name:"is-mutex",value:"0",type:"radio"},domProps:{checked:t._q(t.is_mutex,"0")},on:{change:function(e){t.is_mutex="0"}}}),t._v("取消")])]),t._v(" "),s("div",[s("label",[s("input",{directives:[{name:"model",rawName:"v-model",value:t.is_timeout,expression:"is_timeout"}],attrs:{name:"is-timeout",value:"1",type:"radio"},domProps:{checked:t._q(t.is_timeout,"1")},on:{change:function(e){t.is_timeout="1"}}}),t._v("已过期")]),t._v(" "),s("label",[s("input",{directives:[{name:"model",rawName:"v-model",value:t.is_timeout,expression:"is_timeout"}],attrs:{name:"is-timeout",value:"2",type:"radio"},domProps:{checked:t._q(t.is_timeout,"2")},on:{change:function(e){t.is_timeout="2"}}}),t._v("有效期内")]),t._v(" "),s("label",[s("input",{directives:[{name:"model",rawName:"v-model",value:t.is_timeout,expression:"is_timeout"}],attrs:{name:"is-timeout",value:"0",type:"radio"},domProps:{checked:t._q(t.is_timeout,"0")},on:{change:function(e){t.is_timeout="0"}}}),t._v("取消")])])])]),t._v(" "),s("div",[s("div",{staticClass:"tables"},[s("table",{staticClass:"table table-bordered",attrs:{width:"100%"}},[s("tbody",t._l(t.cron_list,function(e){return s("tr",[s("td",{attrs:{scope:"row"}},[s("div",[s("span",[t._v("id：")]),s("span",[t._v(t._s(e.id))])]),t._v(" "),s("div",[s("span",[t._v("定时配置：")]),s("span",[t._v(t._s(e.cron_set))])]),t._v(" "),s("div",[s("span",[t._v("互斥：")]),t._v(" "),e.is_mutex?s("span",[t._v("是")]):s("span",[t._v("否")])]),t._v(" "),s("div",[s("span",[t._v("运行时间范围：")]),t._v(" "),s("span",[t._v(t._s(e.start_time)+" - "+t._s(e.end_time))])]),t._v(" "),s("div",[s("span",[t._v("执行指令：")]),t._v(" "),s("span",[t._v(t._s(e.command))])]),t._v(" "),s("div",[s("span",[t._v("正在运行：")]),t._v(" "),e.stop?s("span",[s("label",{staticStyle:{color:"#f00","font-weight":"bold"}},[t._v("否")])]):s("span",[t._v("是")])]),t._v(" "),s("div",[s("span",[t._v("并行进程数：")]),t._v(" "),s("span",[t._v(t._s(e.process_num))])]),t._v(" "),s("div",[s("span",[t._v("备注信息：")]),t._v(" "),s("span",[t._v(t._s(e.remark))])]),t._v(" "),s("div",[e.stop?s("a",{staticClass:"btn",attrs:{"item-id":e.id},on:{click:t.start}},[t._v("开始")]):s("a",{staticClass:"btn",attrs:{"item-id":e.id},on:{click:t.stop}},[t._v("停止")]),t._v(" "),s("a",{staticClass:"btn",attrs:{"item-id":e.id},on:{click:t.jumpEdit}},[t._v("编辑")]),t._v(" "),s("a",{staticClass:"btn",attrs:{"item-id":e.id},on:{click:t.del}},[t._v("删除")]),t._v(" "),e.is_mutex?s("a",{staticClass:"btn",attrs:{"item-id":e.id},on:{click:t.mutex0}},[t._v("取消互斥")]):s("a",{staticClass:"btn",attrs:{"item-id":e.id},on:{click:t.mutex1}},[t._v("设为互斥")]),t._v(" "),s("a",{staticClass:"btn",attrs:{"item-id":e.id},on:{click:t.jumpLogs}},[t._v("运行日志")]),t._v(" "),s("a",{staticClass:"btn",attrs:{"item-id":e.id,"item-command":e.command},on:{click:t.run}},[t._v("运行")])])])])}))])])])])])},staticRenderFns:[function(){var t=this.$createElement,e=this._self._c||t;return e("div",{staticClass:"sub-heard-part"},[e("ol",{staticClass:"breadcrumb m-b-0"},[e("li",[e("a",{attrs:{href:"/ui/#/"}},[this._v("首页")])]),this._v(" "),e("li",{staticClass:"active"},[this._v("定时任务管理")])])])}]};var p=function(t){s("iSpo")},f=s("Mz/3")(m.a,h,!1,p,null,null).exports,g=s("Ces+"),x={render:function(){var t=this,e=t.$createElement,s=t._self._c||e;return s("div",[t._m(0),t._v(" "),s("div",{staticClass:"forms-main"},[s("h2",{staticClass:"inner-tittle"},[t._v("编辑定时任务 ")]),t._v(" "),s("div",{staticClass:"graph-form"},[s("div",{staticClass:"form-body"},[s("div",{staticClass:"form-group"},[s("label",{attrs:{for:"cron-set"}},[t._v("定时配置，如：*/1 * * * * *，这里精确到秒，前面的意思是每秒执行一次，分别对应，秒分时日月周")]),t._v(" "),s("input",{directives:[{name:"model",rawName:"v-model",value:t.cron_info.cron_set,expression:"cron_info.cron_set"}],staticClass:"form-control",attrs:{type:"text",id:"cron-set"},domProps:{value:t.cron_info.cron_set,value:t.cron_info.cron_set},on:{input:function(e){e.target.composing||t.$set(t.cron_info,"cron_set",e.target.value)}}})]),t._v(" "),s("div",{staticClass:"form-group"},[s("label",[t._v("开始时间，大于等于此时间才执行，不限留空")]),t._v(" "),s("input",{directives:[{name:"model",rawName:"v-model",value:t.cron_info.start_time,expression:"cron_info.start_time"}],staticClass:"form-control",attrs:{type:"text",id:"start-time"},domProps:{value:t.cron_info.start_time,value:t.cron_info.start_time},on:{input:function(e){e.target.composing||t.$set(t.cron_info,"start_time",e.target.value)}}})]),t._v(" "),s("div",{staticClass:"form-group"},[s("label",{attrs:{for:"end-time"}},[t._v("结束时间，小于此时间才执行，不限留空")]),t._v(" "),s("input",{directives:[{name:"model",rawName:"v-model",value:t.cron_info.end_time,expression:"cron_info.end_time"}],staticClass:"form-control",attrs:{type:"text",id:"end-time"},domProps:{value:t.cron_info.end_time,value:t.cron_info.end_time},on:{input:function(e){e.target.composing||t.$set(t.cron_info,"end_time",e.target.value)}}})]),t._v(" "),s("div",{staticClass:"form-group"},[s("label",{attrs:{for:"command"}},[t._v("执行命令")]),t._v(" "),s("input",{directives:[{name:"model",rawName:"v-model",value:t.cron_info.command,expression:"cron_info.command"}],staticClass:"form-control",attrs:{type:"text",id:"command"},domProps:{value:t.cron_info.command,value:t.cron_info.command},on:{input:function(e){e.target.composing||t.$set(t.cron_info,"command",e.target.value)}}})]),t._v(" "),s("div",{staticClass:"checkbox"},[s("label",[t.cron_info.stop?s("input",{directives:[{name:"model",rawName:"v-model",value:t.cron_info.stop,expression:"cron_info.stop"}],attrs:{type:"checkbox",id:"cron-stop",checked:""},domProps:{checked:Array.isArray(t.cron_info.stop)?t._i(t.cron_info.stop,null)>-1:t.cron_info.stop},on:{change:function(e){var s=t.cron_info.stop,a=e.target,i=!!a.checked;if(Array.isArray(s)){var o=t._i(s,null);a.checked?o<0&&t.$set(t.cron_info,"stop",s.concat([null])):o>-1&&t.$set(t.cron_info,"stop",s.slice(0,o).concat(s.slice(o+1)))}else t.$set(t.cron_info,"stop",i)}}}):s("input",{directives:[{name:"model",rawName:"v-model",value:t.cron_info.stop,expression:"cron_info.stop"}],attrs:{type:"checkbox",id:"cron-stop"},domProps:{checked:Array.isArray(t.cron_info.stop)?t._i(t.cron_info.stop,null)>-1:t.cron_info.stop},on:{change:function(e){var s=t.cron_info.stop,a=e.target,i=!!a.checked;if(Array.isArray(s)){var o=t._i(s,null);a.checked?o<0&&t.$set(t.cron_info,"stop",s.concat([null])):o>-1&&t.$set(t.cron_info,"stop",s.slice(0,o).concat(s.slice(o+1)))}else t.$set(t.cron_info,"stop",i)}}}),t._v("\n            初始化为停止状态\n          ")])]),t._v(" "),s("div",{staticClass:"checkbox"},[s("label",[t.cron_info.is_mutex?s("input",{directives:[{name:"model",rawName:"v-model",value:t.cron_info.is_mutex,expression:"cron_info.is_mutex"}],attrs:{type:"checkbox",id:"cron-is-mutex",checked:""},domProps:{checked:Array.isArray(t.cron_info.is_mutex)?t._i(t.cron_info.is_mutex,null)>-1:t.cron_info.is_mutex},on:{change:function(e){var s=t.cron_info.is_mutex,a=e.target,i=!!a.checked;if(Array.isArray(s)){var o=t._i(s,null);a.checked?o<0&&t.$set(t.cron_info,"is_mutex",s.concat([null])):o>-1&&t.$set(t.cron_info,"is_mutex",s.slice(0,o).concat(s.slice(o+1)))}else t.$set(t.cron_info,"is_mutex",i)}}}):s("input",{directives:[{name:"model",rawName:"v-model",value:t.cron_info.is_mutex,expression:"cron_info.is_mutex"}],attrs:{type:"checkbox",id:"cron-is-mutex"},domProps:{checked:Array.isArray(t.cron_info.is_mutex)?t._i(t.cron_info.is_mutex,null)>-1:t.cron_info.is_mutex},on:{change:function(e){var s=t.cron_info.is_mutex,a=e.target,i=!!a.checked;if(Array.isArray(s)){var o=t._i(s,null);a.checked?o<0&&t.$set(t.cron_info,"is_mutex",s.concat([null])):o>-1&&t.$set(t.cron_info,"is_mutex",s.slice(0,o).concat(s.slice(o+1)))}else t.$set(t.cron_info,"is_mutex",i)}}}),t._v("\n            严格互斥执行\n          ")])]),t._v(" "),s("div",{staticClass:"form-group"},[s("label",{attrs:{for:"remark"}},[t._v("备注")]),t._v(" "),s("textarea",{directives:[{name:"model",rawName:"v-model",value:t.cron_info.remark,expression:"cron_info.remark"}],staticClass:"form-control",attrs:{id:"remark"},domProps:{value:t.cron_info.remark},on:{input:function(e){e.target.composing||t.$set(t.cron_info,"remark",e.target.value)}}},[t._v(t._s(t.cron_info.remark))])]),t._v(" "),s("button",{staticClass:"btn btn-default",attrs:{type:"button",id:"do-submit"},on:{click:t.submit}},[t._v("提交")])])])])])},staticRenderFns:[function(){var t=this.$createElement,e=this._self._c||t;return e("div",{staticClass:"sub-heard-part"},[e("ol",{staticClass:"breadcrumb m-b-0"},[e("li",[e("a",{attrs:{href:"/ui/#/"}},[this._v("首页")])]),this._v(" "),e("li",{staticClass:"active"},[this._v("编辑定时任务")])])])}]},b=s("Mz/3")(g.a,x,!1,null,null,null).exports,y=s("fmqh"),C={render:function(){var t=this,e=t.$createElement,s=t._self._c||e;return s("div",[s("div",[s("label",{staticStyle:{cursor:"pointer"},on:{click:t.searchFailLogs}},[s("input",{directives:[{name:"model",rawName:"v-model",value:t.logs.searchFail,expression:"logs.searchFail"}],attrs:{type:"checkbox"},domProps:{checked:Array.isArray(t.logs.searchFail)?t._i(t.logs.searchFail,null)>-1:t.logs.searchFail},on:{change:function(e){var s=t.logs.searchFail,a=e.target,i=!!a.checked;if(Array.isArray(s)){var o=t._i(s,null);a.checked?o<0&&t.$set(t.logs,"searchFail",s.concat([null])):o>-1&&t.$set(t.logs,"searchFail",s.slice(0,o).concat(s.slice(o+1)))}else t.$set(t.logs,"searchFail",i)}}}),t._v("查看失败记录")]),t._v(" "),s("a",{staticStyle:{cursor:"pointer"},on:{click:t.prevPage}},[t._v("上一页")]),t._v(" "),s("label",[t._v(t._s(t.logs.page)+"/"+t._s(t.logs.totalPage))]),t._v(" "),s("a",{staticStyle:{cursor:"pointer"},on:{click:t.nextPage}},[t._v("下一页")]),t._v("\n    自动刷新 "),s("select",{on:{change:t.refresh}},[s("option",{attrs:{value:"0"}},[t._v("不刷新")]),t._v(" "),s("option",{attrs:{value:"1",selected:""}},[t._v("1s")]),t._v(" "),s("option",{attrs:{value:"5"}},[t._v("5s")]),t._v(" "),s("option",{attrs:{value:"10"}},[t._v("10s")]),t._v(" "),s("option",{attrs:{value:"30"}},[t._v("30s")]),t._v(" "),s("option",{attrs:{value:"60"}},[t._v("60s")])])]),t._v(" "),s("table",{staticClass:"table table-bordered"},[t._m(0),t._v(" "),s("tbody",t._l(t.logs.data,function(e){return s("tr",[s("th",{attrs:{scope:"row"}},[t._v(t._s(e.id))]),t._v(" "),s("th",{attrs:{scope:"row"}},[t._v(t._s(e.cron_id))]),t._v(" "),s("th",{attrs:{scope:"row"}},[t._v(t._s(e.process_id))]),t._v(" "),s("td",[t._v(t._s(e.start_time))]),t._v(" "),s("td",[t._v(t._s(e.state))]),t._v(" "),s("td",[t._v(t._s(e.use_time))]),t._v(" "),s("td",{staticStyle:{"word-break":"break-all"}},[t._v(t._s(e.remark))]),t._v(" "),s("td",{staticStyle:{"word-break":"break-all"}},[t._v(t._s(e.output))]),t._v(" "),s("td",[s("a",{staticClass:"btn",attrs:{"item-process_id":e.process_id,"item-id":e.cron_id},on:{click:t.kill}},[t._v("终止进程")]),t._v(" "),s("a",{staticClass:"btn",attrs:{"item-id":e.id},on:{click:t.detail}},[t._v("详情")])])])}))])])},staticRenderFns:[function(){var t=this,e=t.$createElement,s=t._self._c||e;return s("thead",[s("tr",[s("th",[t._v("#Id")]),t._v(" "),s("th",[t._v("定时任务Id")]),t._v(" "),s("th",[t._v("进程Id")]),t._v(" "),s("th",[t._v("开始执行时间")]),t._v(" "),s("th",[t._v("结果")]),t._v(" "),s("th",[t._v("耗时(毫秒)")]),t._v(" "),s("th",[t._v("备注")]),t._v(" "),s("th",[t._v("输出")]),t._v(" "),s("th",[t._v("操作")])])])}]},k=s("Mz/3")(y.a,C,!1,null,null,null).exports,w={name:"LogDetail",data:function(){return{log_detail:{id:"",cron_id:"",process_id:"",state:"",start_time:"",use_time:0,remark:"",output:""}}},methods:{getInfo:function(){var t=window.location.hash.split("?",-1);console.log(t);var e=this,s={};if(t.length>1){var a=t[1].split("&"),i=0,o=a.length;for(i=0;i<o;i++){var n=a[i].split("=");n.length>1&&(s[n[0]]=n[1])}console.log(s)}void 0===s.id&&(s.id=0),s.id<=0||axios.get("/cron/log/detail/"+s.id).then(function(t){console.log(t),2e3==t.data.code&&(e.log_detail=t.data.data,console.log(e.log))}).catch(function(t){})}},mounted:function(){this.getInfo()}},P={render:function(){var t=this,e=t.$createElement,s=t._self._c||e;return s("div",[s("div",[s("span",{staticClass:"tdis"},[t._v("定时任务ID")]),s("span",[t._v(t._s(t.log_detail.cron_id))])]),t._v(" "),s("div",[s("span",{staticClass:"tdis"},[t._v("进程ID")]),s("span",[t._v(t._s(t.log_detail.process_id))])]),t._v(" "),s("div",[s("span",{staticClass:"tdis"},[t._v("状态")]),s("span",[t._v(t._s(t.log_detail.state))])]),t._v(" "),s("div",[s("span",{staticClass:"tdis"},[t._v("开始时间")]),s("span",[t._v(t._s(t.log_detail.start_time))])]),t._v(" "),s("div",[s("span",{staticClass:"tdis"},[t._v("耗时(毫秒)")]),s("span",[t._v(t._s(t.log_detail.use_time))])]),t._v(" "),s("div",[s("span",{staticClass:"tdis"},[t._v("备注")]),s("span",[t._v(t._s(t.log_detail.remark))])]),t._v(" "),s("div",[s("span",{staticClass:"tdis"},[t._v("输出")]),s("span",[t._v(t._s(t.log_detail.output))])])])},staticRenderFns:[]};var $=s("Mz/3")(w,P,!1,function(t){s("A5SV")},null,null).exports;a.a.use(n.a);var A=new n.a({routes:[{path:"/",name:"Index",component:d},{path:"/add",name:"Add",component:u},{path:"/cron_list",name:"CronList",component:f},{path:"/edit",name:"Edit",component:b},{path:"/logs",name:"Logs",component:k},{path:"/log_detail",name:"LogDetail",component:$}]});a.a.config.productionTip=!1,new a.a({el:"#outter-wp",router:A,components:{App:o},template:"<App/>"})},e1hH:function(t,e){},fmqh:function(t,e,s){"use strict";(function(t){var s=null;e.a={name:"CronList",data:function(){return{logs:{data:[],limit:50,page:1,total:0,totalPage:0,searchFail:!1}}},mounted:function(){window.location.href;var t=this;t.getLogs("mounted"),s=window.setInterval(function(){t.getLogs("refresh setInterval")},1e3)},methods:{searchFailLogs:function(){this.logs.page=1;var t=this;console.log(this.logs.searchFail),window.setTimeout(function(){console.log(t.logs.searchFail),t.getLogs("setTimeout")},20)},getLogs:function(t){console.log(t);var e=window.location.hash.split("?",-1);console.log(e);var s=this,a={};if(e.length>1){var i=e[1].split("&"),o=0,n=i.length;for(o=0;o<n;o++){var r=i[o].split("=");r.length>1&&(a[r[0]]=r[1])}console.log(a)}void 0===a.id&&(a.id=0);var c="0";s.logs.searchFail&&(c="1"),axios.get("/log/list/"+a.id+"/"+c+"/"+s.logs.page+"/"+s.logs.limit).then(function(t){2e3==t.data.code?(console.log(t),s.logs.data=t.data.data.data,s.logs.limit=t.data.data.limit,s.logs.page=t.data.data.page,s.logs.total=t.data.data.total,s.logs.totalPage=t.data.data.totalPage):alert(t.data.message)}).catch(function(t){})},prevPage:function(){var t=this.logs.page-1;t<1&&(t=this.logs.totalPage),this.logs.page=t,this.getLogs("prevPage")},nextPage:function(){var t=this.logs.page+1;t>this.logs.totalPage&&(t=1),this.logs.page=t,this.getLogs("nextPage")},refresh:function(e){var a=this,i=t(e.target).val();console.log(i),null!=s&&window.clearInterval(s),i>0&&(s=window.setInterval(function(){a.getLogs("refresh setInterval")},1e3*i))},kill:function(e){var s=t(e.target).attr("item-id"),a=t(e.target).attr("item-process_id");axios.get("/cron/kill/"+s+"/"+a).then(function(s){console.log(s),2e3==s.data.code?t(e.target).html("kill成功"):t(e.target).html("kill失败"),window.setTimeout(function(){t(e.target).html("终止进程")},3e3)}).catch(function(t){})},detail:function(e){var s=t(e.target).attr("item-id");window.location.href="/ui/#/log_detail?id="+s}}}}).call(e,s("ZosL"))},iSpo:function(t,e){}},["NHnr"]);
//# sourceMappingURL=app.e1e6d6444ed45eafbeda.js.map