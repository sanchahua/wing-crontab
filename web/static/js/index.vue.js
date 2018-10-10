// Calculator for converting Miles to Kilometers



$(document).ready(function() {
  var needle = $("#needle");

  TweenLite.to(needle, 2, {rotation:-31,  transformOrigin:"bottom right"});

  // select current content in input boxes on click
  $("input[type='text']").on("click", function () {
    $(this).select();
  });

  //clear kilometers value when miles is selected
  $("#miles").focus(function(){
    $("#kilometers").val('');
  });

  //clear miles value when kilometers is selected
  $("#kilometers").focus(function(){
    $("#miles").val('');
  });

  // convert miles to kilometers
  $('#miles').keyup(function() {
    var mi = $(this).val();
    var	miNum = parseInt(mi) * 1.6093;
    //make sure kmNum is a number then output
    if ( (mi <= 75) && !isNaN(miNum) ){
      var speedMi = miNum * 2 - 31;
      $('#numbers').css('text-align', 'center');
      $('#kilometers').val(miNum.toFixed(2));
      $('#numbers').html(miNum.toFixed(0));
      $('#mi-km').html('Kilometers');
    } else if (!isNaN(miNum)){
      var speedMi = 215;
      $('#numbers').css('text-align', 'right');
      $('#kilometers').val(miNum.toFixed(2));
      $('#numbers').html(miNum.toFixed(0));
      $('#mi-km').html('Kilometers');
    } else {
      $('#miles').val('');
      $('#kilometers').val('');
      $('#numbers').html('');
      $("#errmsg").html("Numbers Only").show().fadeOut(1600);
    }

    var needle = $("#needle");
    TweenLite.to(needle, 2, {rotation:speedMi,  transformOrigin:"bottom right"});
  });

  // convert kilometers to miles
  $('#kilometers').keyup(function() {
    var km = $(this).val();
    var	kmNum = parseInt(km) * 0.62137;
    //make sure kmNum is a number then output
    if ( (km <= 195) && !isNaN(kmNum) ){
      var speedKm = kmNum * 2 - 31;
      $('#numbers').css('text-align', 'center');
      $('#miles').val(kmNum.toFixed(2));
      $('#numbers').html(kmNum.toFixed(0));
      $('#mi-km').html('Miles');
    } else if (!isNaN(kmNum)){
      var speedKm= 215;
      $('#numbers').css('text-align', 'right');
      $('#miles').val(kmNum.toFixed(2));
      $('#numbers').html(kmNum.toFixed(0));
      $('#mi-km').html('Miles');
    } else {
      $('#miles').val('');
      $('#kilometers').val('');
      $('#numbers').html('');
      $("#errmsg").html("Numbers Only").show().fadeOut(1600);
    }

    var needle = $("#needle");
    TweenLite.to(needle, 2, {rotation:speedKm,  transformOrigin:"bottom right"});
  });

  $(document).ready(function () {
    data = {
      '2010' : 300,
      '2011' : 200,
      '2012' : 100,
      '2013' : 500,
      '2014' : 400,
      '2015' : 200
    };

    $("#chart1").faBoChart({
      time: 500,
      animate: true,
      instantAnimate: true,
      straight: false,
      data: data,
      labelTextColor : "#002561",
    });
    // $("#chart2").faBoChart({
    //   time: 2500,
    //   animate: true,
    //   data: data,
    //   straight: true,
    //   labelTextColor : "#002561",
    // });
  });
  // var chart = AmCharts.makeChart("chartdiv1", {
  //   "type": "serial",
  //   "theme": "light",
  //   "rotate": true,
  //   "marginBottom": 50,
  //   "dataProvider": [{
  //     "age": "85+",
  //     "male": -0.1,
  //     "female": 0.3
  //   }, {
  //     "age": "80-54",
  //     "male": -0.2,
  //     "female": 0.3
  //   }, {
  //     "age": "75-79",
  //     "male": -0.3,
  //     "female": 0.6
  //   }, {
  //     "age": "70-74",
  //     "male": -0.5,
  //     "female": 0.8
  //   }, {
  //     "age": "65-69",
  //     "male": -0.8,
  //     "female": 1.0
  //   }, {
  //     "age": "60-64",
  //     "male": -1.1,
  //     "female": 1.3
  //   }, {
  //     "age": "55-59",
  //     "male": -1.7,
  //     "female": 1.9
  //   }, {
  //     "age": "50-54",
  //     "male": -2.2,
  //     "female": 2.5
  //   }, {
  //     "age": "45-49",
  //     "male": -2.8,
  //     "female": 3.0
  //   }, {
  //     "age": "40-44",
  //     "male": -3.4,
  //     "female": 3.6
  //   }, {
  //     "age": "35-39",
  //     "male": -4.2,
  //     "female": 4.1
  //   }, {
  //     "age": "30-34",
  //     "male": -5.2,
  //     "female": 4.8
  //   }, {
  //     "age": "25-29",
  //     "male": -5.6,
  //     "female": 5.1
  //   }, {
  //     "age": "20-24",
  //     "male": -5.1,
  //     "female": 5.1
  //   }, {
  //     "age": "15-19",
  //     "male": -3.8,
  //     "female": 3.8
  //   }, {
  //     "age": "10-14",
  //     "male": -3.2,
  //     "female": 3.4
  //   }, {
  //     "age": "5-9",
  //     "male": -4.4,
  //     "female": 4.1
  //   }, {
  //     "age": "0-4",
  //     "male": -5.0,
  //     "female": 4.8
  //   }],
  //   "startDuration": 1,
  //   "graphs": [{
  //     "fillAlphas": 0.8,
  //     "lineAlpha": 0.2,
  //     "type": "column",
  //     "valueField": "male",
  //     "title": "Male",
  //     "labelText": "[[value]]",
  //     "clustered": false,
  //     "labelFunction": function(item) {
  //       return Math.abs(item.values.value);
  //     },
  //     "balloonFunction": function(item) {
  //       return item.category + ": " + Math.abs(item.values.value) + "%";
  //     }
  //   }, {
  //     "fillAlphas": 0.8,
  //     "lineAlpha": 0.2,
  //     "type": "column",
  //     "valueField": "female",
  //     "title": "Female",
  //     "labelText": "[[value]]",
  //     "clustered": false,
  //     "labelFunction": function(item) {
  //       return Math.abs(item.values.value);
  //     },
  //     "balloonFunction": function(item) {
  //       return item.category + ": " + Math.abs(item.values.value) + "%";
  //     }
  //   }],
  //   "categoryField": "age",
  //   "categoryAxis": {
  //     "gridPosition": "start",
  //     "gridAlpha": 0.2,
  //     "axisAlpha": 0
  //   },
  //   "valueAxes": [{
  //     "gridAlpha": 0,
  //     "ignoreAxisWidth": true,
  //     "labelFunction": function(value) {
  //       return Math.abs(value) + '%';
  //     },
  //     "guides": [{
  //       "value": 0,
  //       "lineAlpha": 0.2
  //     }]
  //   }],
  //   "balloon": {
  //     "fixedPosition": true
  //   },
  //   "chartCursor": {
  //     "valueBalloonsEnabled": false,
  //     "cursorAlpha": 0.05,
  //     "fullWidth": true
  //   },
  //   "allLabels": [{
  //     "text": "Male",
  //     "x": "28%",
  //     "y": "97%",
  //     "bold": true,
  //     "align": "middle"
  //   }, {
  //     "text": "Female",
  //     "x": "75%",
  //     "y": "97%",
  //     "bold": true,
  //     "align": "middle"
  //   }],
  //   "export": {
  //     "enabled": true
  //   }
  //
  // });


  // var chart = AmCharts.makeChart( "chartdiv2", {
  //   "type": "serial",
  //   "theme": "patterns",
  //   "legend": {
  //     "useGraphSettings": true
  //   },
  //   "dataProvider": [{
  //     "year": 1930,
  //     "italy": 1,
  //     "germany": 5,
  //     "uk": 3
  //   }, {
  //     "year": 1934,
  //     "italy": 1,
  //     "germany": 2,
  //     "uk": 6
  //   }, {
  //     "year": 1938,
  //     "italy": 2,
  //     "germany": 3,
  //     "uk": 1
  //   }, {
  //     "year": 1950,
  //     "italy": 3,
  //     "germany": 4,
  //     "uk": 1
  //   }, {
  //     "year": 1954,
  //     "italy": 5,
  //     "germany": 1,
  //     "uk": 2
  //   }, {
  //     "year": 1958,
  //     "italy": 3,
  //     "germany": 2,
  //     "uk": 1
  //   }, {
  //     "year": 1962,
  //     "italy": 1,
  //     "germany": 2,
  //     "uk": 3
  //   }, {
  //     "year": 1966,
  //     "italy": 2,
  //     "germany": 1,
  //     "uk": 5
  //   }, {
  //     "year": 1970,
  //     "italy": 3,
  //     "germany": 5,
  //     "uk": 2
  //   }, {
  //     "year": 1974,
  //     "italy": 4,
  //     "germany": 3,
  //     "uk": 6
  //   }, {
  //     "year": 1978,
  //     "italy": 1,
  //     "germany": 2,
  //     "uk": 4
  //   }],
  //   "valueAxes": [{
  //     "integersOnly": true,
  //     "maximum": 6,
  //     "minimum": 1,
  //     "reversed": true,
  //     "axisAlpha": 0,
  //     "dashLength": 5,
  //     "gridCount": 10,
  //     "position": "left",
  //     "title": "Place taken"
  //   }],
  //   "startDuration": 0.5,
  //   "graphs": [{
  //     "balloonText": "place taken by Italy in [[category]]: [[value]]",
  //     "bullet": "round",
  //     "hidden": true,
  //     "title": "Italy",
  //     "valueField": "italy",
  //     "fillAlphas": 0
  //   }, {
  //     "balloonText": "place taken by Germany in [[category]]: [[value]]",
  //     "bullet": "round",
  //     "title": "Germany",
  //     "valueField": "germany",
  //     "fillAlphas": 0
  //   }, {
  //     "balloonText": "place taken by UK in [[category]]: [[value]]",
  //     "bullet": "round",
  //     "title": "United Kingdom",
  //     "valueField": "uk",
  //     "fillAlphas": 0
  //   }],
  //   "chartCursor": {
  //     "cursorAlpha": 0,
  //     "zoomable": false
  //   },
  //   "categoryField": "year",
  //   "categoryAxis": {
  //     "gridPosition": "start",
  //     "axisAlpha": 0,
  //     "fillAlpha": 0.05,
  //     "fillColor": "#000000",
  //     "gridAlpha": 0,
  //     "position": "top"
  //   },
  //   "export": {
  //     "enabled": true,
  //     "position": "bottom-right"
  //   }
  // });
  // var chart = AmCharts.makeChart( "chartdiv4", {
  //   "type": "radar",
  //   "theme": "light",
  //   "dataProvider": [ {
  //     "direction": "N",
  //     "value": 8
  //   }, {
  //     "direction": "NE",
  //     "value": 9
  //   }, {
  //     "direction": "E",
  //     "value": 4.5
  //   }, {
  //     "direction": "SE",
  //     "value": 3.5
  //   }, {
  //     "direction": "S",
  //     "value": 9.2
  //   }, {
  //     "direction": "SW",
  //     "value": 8.4
  //   }, {
  //     "direction": "W",
  //     "value": 11.1
  //   }, {
  //     "direction": "NW",
  //     "value": 10
  //   } ],
  //   "valueAxes": [ {
  //     "gridType": "circles",
  //     "minimum": 0,
  //     "autoGridCount": false,
  //     "axisAlpha": 0.2,
  //     "fillAlpha": 0.05,
  //     "fillColor": "#FFFFFF",
  //     "gridAlpha": 0.08,
  //     "guides": [ {
  //       "angle": 225,
  //       "fillAlpha": 0.7,
  //       "fillColor": "#052963",
  //       "tickLength": 0,
  //       "toAngle": 315,
  //       "toValue": 14,
  //       "value": 0,
  //       "lineAlpha": 0,
  //
  //     }, {
  //       "angle": 45,
  //       "fillAlpha": 0.6,
  //       "fillColor": "#ea4c89",
  //       "tickLength": 0,
  //       "toAngle": 135,
  //       "toValue": 14,
  //       "value": 0,
  //       "lineAlpha": 0,
  //     } ],
  //     "position": "left"
  //   } ],
  //   "startDuration": 1,
  //   "graphs": [ {
  //     "balloonText": "[[category]]: [[value]] m/s",
  //     "bullet": "round",
  //     "fillAlphas": 0.3,
  //     "valueField": "value"
  //   } ],
  //   "categoryField": "direction",
  //   "export": {
  //     "enabled": true
  //   }
  // } );

  var randomScalingFactor = function() {
    return Math.round(Math.random() * 100 * (Math.random() > 0.5 ? -1 : 1));
  };
  var randomColor = function(opacity) {
    return 'rgba(' + Math.round(Math.random() * 255) + ',' + Math.round(Math.random() * 255) + ',' + Math.round(Math.random() * 255) + ',' + (opacity || '.3') + ')';
  };

  var lineChartData = {
    labels: ["January", "February", "March", "April", "May", "June", "July"],
    datasets: [{
      label: "My First dataset",
      data: [randomScalingFactor(), randomScalingFactor(), randomScalingFactor(), randomScalingFactor(), randomScalingFactor(), randomScalingFactor(), randomScalingFactor()],
      yAxisID: "y-axis-1",
    }, {
      label: "My Second dataset",
      data: [randomScalingFactor(), randomScalingFactor(), randomScalingFactor(), randomScalingFactor(), randomScalingFactor(), randomScalingFactor(), randomScalingFactor()],
      yAxisID: "y-axis-2"
    }]
  };

  $.each(lineChartData.datasets, function(i, dataset) {
    dataset.borderColor = randomColor(0.4);
    dataset.backgroundColor = randomColor(1);
    dataset.pointBorderColor = randomColor(0.7);
    dataset.pointBackgroundColor = randomColor(0.5);
    dataset.pointBorderWidth = 1;
  });

  console.log(lineChartData);

  window.onload = function() {
    // var ctx = document.getElementById("canvas").getContext("2d");
    // window.myLine = Chart.Line(ctx, {
    //   data: lineChartData,
    //   options: {
    //
    //     hoverMode: 'label',
    //     stacked: false,
    //     scales: {
    //       xAxes: [{
    //         display: true,
    //         gridLines: {
    //           offsetGridLines: false
    //         }
    //       }],
    //       yAxes: [{
    //         type: "linear", // only linear but allow scale type registration. This allows extensions to exist solely for log scale for instance
    //         display: true,
    //         position: "left",
    //         id: "y-axis-1",
    //       }, {
    //         type: "linear", // only linear but allow scale type registration. This allows extensions to exist solely for log scale for instance
    //         display: true,
    //         position: "right",
    //         id: "y-axis-2",
    //
    //         // grid line settings
    //         gridLines: {
    //           drawOnChartArea: false, // only want the grid lines for one axis to show up
    //         },
    //       }],
    //     }
    //   }
    // });
  };

  $('#randomizeData').click(function() {
    lineChartData.datasets[0].data = [randomScalingFactor(), randomScalingFactor(), randomScalingFactor(), randomScalingFactor(), randomScalingFactor(), randomScalingFactor(), randomScalingFactor()];

    lineChartData.datasets[1].data = [randomScalingFactor(), randomScalingFactor(), randomScalingFactor(), randomScalingFactor(), randomScalingFactor(), randomScalingFactor(), randomScalingFactor()];

    window.myLine.update();
  });

  var icons = new Skycons({"color": "#002561"}),
    list  = [
      "clear-night", "partly-cloudy-day",
      "partly-cloudy-night", "cloudy", "rain", "sleet", "snow", "wind",
      "fog"
    ],
    i;

  for(i = list.length; i--; )
    icons.set(list[i], list[i]);

  icons.play();

  var icons = new Skycons({"color": "#00C6D7"}),
    list  = [
      "clear-night", "cloudy",
      "partly-cloudy-night", "cloudy", "rain", "sleet", "snow", "wind",
      "fog"
    ],
    i;

  for(i = list.length; i--; )
    icons.set(list[i], list[i]);

  icons.play();

  var icons = new Skycons({"color": "#00C6D7"}),
    list  = [
      "clear-night", "clear-day",
      "partly-cloudy-night", "cloudy", "rain", "sleet", "snow", "wind",
      "fog"
    ],
    i;

  for(i = list.length; i--; )
    icons.set(list[i], list[i]);

  icons.play();


  var icons = new Skycons({"color": "#00C6D7"}),
    list  = [
      "clear-night", "clear-day",
      "partly-cloudy-night", "cloudy", "rain", "sleet", "snow", "wind",
      "fog"
    ],
    i;

  for(i = list.length; i--; )
    icons.set(list[i], list[i]);

  icons.play();

  var icons = new Skycons({"color": "#00C6D7"}),
    list  = [
      "clear-night", "clear-day",
      "partly-cloudy-night", "cloudy", "rain", "sleet", "snow", "wind",
      "fog"
    ],
    i;

  for(i = list.length; i--; )
    icons.set(list[i], list[i]);

  icons.play();
});
