(function(){
'use strict';
var $=function(id){return document.getElementById(id)}, state={}, dragging=null, animation=false;
function icon(el,name){el.innerHTML=window.dashboardIcons[name];}
document.querySelectorAll('[data-icon]').forEach(function(el){icon(el,el.dataset.icon)});
function setSlider(id,value){
  value=Math.max(0,Math.min(100,Math.round(value)));
  state[id]=App.command(id==='volume'?'audio.volume':'playback.seek',value);
  var el=$(id);el.dataset.value=state[id];el.querySelector('.fill').style.width=state[id]+'%';el.querySelector('.thumb').style.left=state[id]+'%';
  if(id==='volume')$('volume-value').textContent=state[id];else timing();
}
function slide(e){if(!dragging)return;if(dragging==='seek')state.duration=180;var rect=$(dragging).getBoundingClientRect();setSlider(dragging,(e.clientX/(parseFloat(document.body.style.zoom)||1)-rect.left)/rect.width*100);e.preventDefault();}
['volume','seek'].forEach(function(id){$(id).onmousedown=function(e){dragging=id;slide(e);};});
document.addEventListener('mousemove',slide);document.addEventListener('mouseup',function(){dragging=null;});
function dropdown(id,code,options){var el=$(id);options.forEach(function(label,index){var row=document.createElement('div');row.className='option';row.textContent=label;row.onclick=function(e){e.stopPropagation();App.command(code,index);el.querySelector('.selection').textContent=label;el.classList.remove('open');};el.querySelector('.options').appendChild(row);});el.onclick=function(e){e.stopPropagation();var was=el.classList.contains('open');close();if(!was)el.classList.add('open');};}
function close(){var open=document.querySelectorAll('.dropdown.open');open.forEach(function(el){el.classList.remove('open')});return open.length>0;}
dropdown('output-device','audio.output',['Internal Speakers','Headphones','USB Audio']);dropdown('input-device','audio.input',['Internal Microphone','USB Microphone','Line In']);document.addEventListener('mousedown',function(e){if(!e.target.closest('.dropdown'))close();});
function playing(value){state.playing=!!App.command('playback.playing',value?1:0);icon($('play'),state.playing?'pause-fill':'play-fill');if(state.playing)state.duration=180;timing();}
$('play').onclick=function(){playing(!state.playing)};$('previous').onclick=function(){setSlider('seek',state.seek-10)};$('next').onclick=function(){setSlider('seek',state.seek+10)};
['tone','record','analyzer'].forEach(function(id){$(id).onclick=function(){state[id]=!!App.command('tools.'+id,state[id]?0:1);$(id).classList.toggle('active',state[id]);};});
$('mute').onclick=function(){state.mute=!state.mute;App.command('audio.mute',state.mute?1:0);icon($('mute'),state.mute?'volume-mute-fill':'volume-up-fill');};
function time(s){return ('0'+Math.floor(s/60)).slice(-2)+':'+('0'+Math.floor(s%60)).slice(-2)}
function timing(){var duration=state.duration||0;$('elapsed').textContent=time(duration*state.seek/100)+' / '+time(duration);}
for(var i=0;i<32;i++){var led=document.createElement('i');led.className=i<16?'green':i>=27&&i<30?'orange':'';$('meter').appendChild(led);}
var heights=[28,40,44,48,43,58,67,77,70,64,56,53,46,43,37,47,38,30,33,34,31,38,31,25,23,20,25,20,18,15,17,25,19,17,15,12,11,10,11,9,8,6,8,7,6,5,4,3];
heights.forEach(function(h){var bar=document.createElement('i');bar.style.height=h+'%';$('bars').appendChild(bar)});
function reset(){state={duration:0,playing:false,mute:false,tone:false,record:false,analyzer:false};setSlider('volume',80);setSlider('seek',0);playing(false);icon($('mute'),'volume-up-fill');App.command('audio.mute',0);['tone','record','analyzer'].forEach(function(id){$(id).classList.remove('active');App.command('tools.'+id,0)});close();}
window.dashboard={reset:reset,escape:function(){if(modals.escape())return true;var panel=popovers.close();return close()||panel;},animate:function(value){animation=value},tick:function(clock){$('clock').textContent=clock},state: function(){return JSON.stringify(state)}};
reset();var last=Date.now();setInterval(function(){var now=Date.now(),delta=(now-last)/1000;last=now;if(state.playing&&!dragging){state.seek=Math.min(100,state.seek+delta/180*100);$('seek').querySelector('.thumb').style.left=state.seek+'%';$('seek').querySelector('.fill').style.width=state.seek+'%';timing();if(state.seek>=100)playing(false);}if(animation||state.analyzer||state.playing){var bars=$('bars').children;heights.forEach(function(h,i){bars[i].style.height=Math.max(2,h+Math.sin(now/350+i*.7)*5)+'%';});}},50);
})();
