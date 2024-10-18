function populate(canvasid){
  const canvas = document.getElementById(canvasid);
  const engine = new BABYLON.Engine(canvas, true);
  
  canvas.onwheel = function(event){
    event.preventDefault();
  };
  
  var scene = new BABYLON.Scene(engine);
  scene.clearColor = new BABYLON.Color3.FromHexString("#28263d")
  // var helper = scene.createDefaultEnvironment({
  //   // createSkybox: false,
  // });
  //Adding a light
  // var light = new BABYLON.PointLight("Omni", new BABYLON.Vector3(20, 20, 100), scene);
  
  //Adding an Arc Rotate Camera
  // var camera = new BABYLON.ArcRotateCamera("Camera", 0, 0, 0, BABYLON.Vector3.Zero(), scene);
  // var camera = new BABYLON.UniversalCamera("UniversalCamera", new BABYLON.Vector3(0, 0, -5), scene);
  const alpha = Math.PI/4;
  const beta = Math.PI/3;
  const radius = 2.5;
  const target = new BABYLON.Vector3(0, .8, 0);
  
  const camera = new BABYLON.ArcRotateCamera("Camera", alpha, beta, radius, target, scene);
  camera.attachControl(canvas, true);
  camera.wheelPrecision = 100
  camera.useFramingBehavior = true
  camera._panningMouseButton = 1
  
  const light = new BABYLON.HemisphericLight("light", new BABYLON.Vector3(1, 1, 1));
  
  // The first parameter can be used to specify which mesh to import. Here we import all meshes
  BABYLON.SceneLoader.ImportMesh("", "./", `./Animations/${canvasid}.glb`, scene, function (newMeshes) {
      // Set the target of the camera to the first imported mesh
      // camera.parent = newMeshes[0];
      camera.setTarget(newMeshes[1]);
  });
  
  // Move the light with the camera
  // scene.registerBeforeRender(function () {
  //     light.position = camera.position;
  // });
  
  engine.runRenderLoop(function(){
    scene.render();
  });
}

let elements = document.getElementsByTagName("canvas");

for (let element of elements){
  console.log(element.id)
  populate(element.id)
};