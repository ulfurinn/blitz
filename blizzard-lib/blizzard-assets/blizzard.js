"use strict";

var app = angular.module("blizzard", []);

app.controller("BlizzardCtl", ["$scope", function($scope){

    $scope.wsOn = false;
    $scope.routes = [];
    $scope.procGroups = {};
    $scope.apps = {};

    $scope.addRoute = function(route) {
        $scope.routes.push(route.pathSpec);
    };

    $scope.updateApp = function(app) {
        if( $scope.apps[app.id] ) {
            for( var p in app ) {
                $scope.apps[app.id][p] = app[p];
            }
        } else {
            $scope.apps[app.id] = app;
        }
    };

    $scope.deleteApp = function(app) {
        delete $scope.apps[app.id];
    };

    $scope.updateProcGroup = function(pg) {
        if( $scope.procGroups[pg.id] ) {
            for(var p in pg) {
                $scope.procGroups[pg.id][p] = pg[p];
            }
        } else {
            pg['procs'] = { };
            $scope.procGroups[pg.id] = pg;
        }
    };

    $scope.deleteProcGroup = function(pg) {
        delete $scope.procGroups[pg.id];
    };

    $scope.updateProc = function(i){
        var pg = $scope.procGroups[i.group];
        if( !pg ) return;
        if( pg.procs[i.id] ) {
            for(var p in i) {
                pg.procs[i.id][p] = i[p];
            }
        } else {
            pg.procs[i.id] = i;
        }
    };

    $scope.deleteProc = function(i){
        var pg = $scope.procGroups[i.group];
        if( !pg ) return;
        delete pg.procs[i.id];
    };

    var ws_onclose, ws_onopen, ws_onmessage;
    var ws_conn;

    var ws_connect = function(){
        console.log("connecting");
        ws_conn = new WebSocket("ws://" + window.location.host + "/ws");
        ws_conn.onclose = ws_onclose;
        ws_conn.onopen = ws_onopen;
        ws_conn.onmessage = ws_onmessage;
    };

    ws_onclose = function(e){
        console.log("closed");
        $scope.$apply(function(){
            $scope.wsOn = false;
            $scope.routes = [];
            $scope.procGroups = {};
        });
        setTimeout(ws_connect, 5000);
    };

    ws_onopen = function(e){
        console.log("open");
        $scope.$apply(function(){
            $scope.wsOn = true;
        });
    };

    ws_onmessage = function(e){
        console.log(e.data);
        var data = JSON.parse(e.data);
        console.log(data);
        $scope.$apply(function(){
            switch(data.type) {
                case "add-route":
                    $scope.addRoute(data);
                    break;
                case "app":
                    $scope.updateApp(data);
                    break;
                case "app-dispose":
                    $scope.deleteApp(data);
                    break;
                case "proc-group":
                    $scope.updateProcGroup(data);
                    break;
                case "proc-group-dispose":
                    $scope.deleteProcGroup(data);
                    break;
                case "proc":
                    $scope.updateProc(data);
                    break;
                case "proc-dispose":
                    $scope.deleteProc(data);
                    break;
            }
        })
    };

    ws_connect();

}]);
