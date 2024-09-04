
var details = document.getElementsByTagName("details");
console.log(details.length);

var expandButtons = document.getElementsByClassName("expand");
console.log(expandButtons.length);

var collapseButtons = document.getElementsByClassName("collapse");
console.log(collapseButtons.length);

var ea = function expand_all(){
  for (var i = 0; i < details.length; i++) {
    details[i].open = true;
  };
};

var ca = function collapse_all(){
  for (var i = 0; i < details.length; i++) {
    details[i].open = false;
  };
};

for (var i = 0; i < expandButtons.length; i++) {
  expandButtons[i].addEventListener("click", ea);
};

for (var i = 0; i < collapseButtons.length; i++) {
  collapseButtons[i].addEventListener("click", ca);
};

ea();
