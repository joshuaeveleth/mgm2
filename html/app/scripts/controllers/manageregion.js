'use strict';

/**
 * @ngdoc function
 * @name mgmApp.controller:ManageregionCtrl
 * @description
 * # ManageregionCtrl
 * Controller of the mgmApp
 */
angular.module('mgmApp')
  .controller('ManageregionCtrl', function($scope, $modalInstance, $timeout, mgm, region) {

    $scope.region = region;
    $scope.estates = [];
    $scope.hosts = [];


    $scope.currentHost = '';
    $scope.currentEstate = '';
    $scope.currentX = region.LocX;
    $scope.currentY = region.LocY;

    function init() {

      $scope.estates = [];
      for (var e in mgm.estates){
        $scope.estates.push(mgm.estates[e]);
      }

      $scope.hosts = [];
      $scope.hosts[0] = {
        ID: 0,
        Hostname: '<none>'
      };
      for (var e in mgm.hosts){
        $scope.hosts[e] = mgm.hosts[e];
      }

      for (var id in $scope.estates) {
        for (var i = 0; i < $scope.estates[id].Regions.length; i++) {
          if ($scope.estates[id].Regions[i] === $scope.region.UUID) {
            $scope.currentEstate = $scope.estates[id];
            $scope.region.Estate = $scope.estates[id].Name;
          }
        }
      }

      for (var e in $scope.hosts) {
        if(e == $scope.region.Host){
          $scope.currentHost = $scope.hosts[e];
        }
      }
      //$scope.currentHost = $scope.hosts[$scope.region.Host];
    }

    $scope.close = function() {
      $modalInstance.close();
    };

    $scope.setXY = function() {
      if ($scope.currentX === undefined) {
        alertify.error('X coordinate is invalid');
        return;
      }
      if ($scope.currentY === undefined) {
        alertify.error('Y coordinate is invalid');
        return;
      }
      $scope.currentX = parseInt($scope.currentX);
      $scope.currentY = parseInt($scope.currentY);
      if ($scope.currentX !== $scope.region.LocX || $scope.currentY !== $scope.region.LocY) {
        //we have a valid location that has changed, make sure we are not stomping on another region
        /*for(var id in mgm.regions){
          if($scope.region.UUID == id){
            //do not compare the region to its-self
            continue;
          }
          var target = mgm.regions[id];

          //check for existing region origin within the bounds of the new location
          var diffX = target.LocX - $scope.currentX;
          var diffY = target.LocY - $scope.currentY;
          if(diffX >= 0 && diffX < region.Size && diffY >= 0 && diffY < region.Size){
            alertify.error('Error: New location overlaps with region ' + target.Name);
            return
          }

          //check for new origin within the bounds of the existing region
          diffX = $scope.currentX - target.LocX;
          diffY = $scope.currentY - target.LocY;
          if(diffX >= 0 && diffX < mgm.regions[id].Size && diffY >= 0 && diffY < target.Size){
            alertify.error('Error: New location overlaps with region ' + target.Name);
            return
          }
        }*/

        //make the mgm request
        mgm.request('SetLocation', {
          'RegionUUID': region.UUID,
          'X': $scope.currentX,
          'Y': $scope.currentY
        }, function(success, msg){
          alertify.log('SetLocation: ' + success + ': ' + msg);
        });
      } else {
        console.log('Subbornly refusing to update XY coordinates that haven\'t changed');
      }
    };

    $scope.setEstate = function() {
      if ($scope.currentEstate.Name !== $scope.region.Estate) {
        mgm.request('SetEstate', {
          'RegionUUID': region.UUID,
          'ID': $scope.currentEstate.ID
        }, function(success, msg) {
          alertify.log('' + success + ': ' + msg);
        });
      } else {
        console.log('Subbornly refusing to update Estate that hasn\'t changed');
      }
    };

    $scope.setHost = function() {
      console.log('change host for ' + $scope.region.Name + ' from ' + $scope.region.Host + ' to ' + $scope.currentHost.ID);
      if ($scope.currentHost.ID !== $scope.region.Host) {
        mgm.request('SetHost', {
          'RegionUUID': region.UUID,
          'ID': $scope.currentHost.ID
        }, function(success, msg){
          alertify.log('' + success +': ' + msg);
        });
      } else {
        console.log('Subbornly refusing to update Host that hasn\'t changed');
      }
    };

    $scope.$on('EstateUpdate', function() {
      $timeout(function() {
        init();
      });
    });

    $scope.$on('RegionUpdate', function(event, r) {
      if (region.UUID === r.UUID) {
        $timeout(function() {
          $scope.region = r;
          init();
        });
      }
    });

    init();
  });
