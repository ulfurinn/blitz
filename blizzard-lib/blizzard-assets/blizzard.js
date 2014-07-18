"use strict";

var app = angular.module("blizzard", []);

app.controller("BlizzardCtl", ["$scope", function($scope){

    $scope.wsOn = false;
    $scope.routes = [];
    $scope.procGroups = [];

    $scope.addRoute = function(route) {
        $scope.routes.push(route);
    }

    $scope.addProcGroup = function(pg) {
        $scope.procGroups.push(pg);
    }

    var conn = new WebSocket("ws://" + window.location.host + "/ws");
    conn.onclose = function(e){
        console.log("closed");
        $scope.$apply(function(){
            $scope.wsOn = false;
        });
    }

    conn.onopen = function(e){
        console.log("open");
        $scope.$apply(function(){
            $scope.wsOn = true;
        });
    }

    conn.onmessage = function(e){
        console.log(e.data);
        var data = JSON.parse(e.data);
        console.log(data);
        $scope.$apply(function(){
            switch(data.type) {
                case "add-route":
                    $scope.addRoute(data.data);
                    break;
                case "add-proc-group":
                    $scope.addProcGroup(data.data);
                    break;
            }
        })
    }

}]);
