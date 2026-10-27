(function(){
  'use strict';
  var $=function(id){return document.getElementById(id)}, current=null, monthKey='', dialog=null;
  var months=['January','February','March','April','May','June','July','August','September','October','November','December'];
  var fields=[['year','Year',2000,2099],['month','Month',1,12],['day','Day',1,31],['hour','Hour',0,23],['minute','Minute',0,59],['second','Second',0,59]];
  function pad(value){return ('0'+value).slice(-2)}
  function update(data){
    current=data;
    var key=data.year+'-'+data.month+'-'+data.day;
    if(monthKey===key)return;
    monthKey=key;$('calendar-month').textContent=months[data.month-1]+' '+data.year;
    var grid=$('calendar-days');grid.innerHTML='';
    var start=new Date(Date.UTC(data.year,data.month-1,1)).getUTCDay();
    var days=new Date(Date.UTC(data.year,data.month,0)).getUTCDate();
    for(var i=0;i<Math.ceil((start+days)/7)*7;i++){var cell=document.createElement('div');var day=i-start+1;cell.className='calendar-day';if(day>0&&day<=days){cell.textContent=day;if(day===data.day)cell.classList.add('today')}grid.appendChild(cell);}
  }
  fields.forEach(function(field,index){
    var group=document.createElement('div');group.className='date-field';
    var label=document.createElement('div');label.className='date-field-label';label.textContent=field[1];group.appendChild(label);
    var box=document.createElement('div');box.className='date-value-box';
    var value=document.createElement('div');value.id='date-'+field[0];value.className='date-value';value.contentEditable='true';value.spellcheck=false;
    value.onfocus=function(){var selection=window.getSelection(),range=document.createRange();range.selectNodeContents(value);selection.removeAllRanges();selection.addRange(range)};
    value.onkeydown=function(e){if(e.key==='Enter'){e.preventDefault();value.blur()}if(e.key==='ArrowUp'||e.key==='ArrowDown'){e.preventDefault();step(e.key==='ArrowUp'?1:-1)}};
    value.oninput=function(){$('date-error').textContent=''};
    box.appendChild(value);
    var arrows=document.createElement('div');arrows.className='date-step-arrows';
    function step(delta){var n=parseInt(value.textContent,10);n=Math.max(field[2],Math.min(field[3],(isNaN(n)?field[2]:n)+delta));value.textContent=index===0?n:pad(n);$('date-error').textContent=''}
    [1,-1].forEach(function(delta){var arrow=document.createElement('div');arrow.innerHTML='<svg viewBox="0 0 12 8"><path d="'+(delta===1?'M1 7L6 1L11 7Z':'M1 1L6 7L11 1Z')+'"/></svg>';arrow.onmousedown=function(e){e.preventDefault()};arrow.onclick=function(){step(delta)};arrows.appendChild(arrow)});
    box.appendChild(arrows);group.appendChild(box);$(index<3?'date-fields':'time-fields').appendChild(group);
  });
  function open(){
    if(!current)return;
    fields.forEach(function(field,index){$('date-'+field[0]).textContent=index===0?current[field[0]]:pad(current[field[0]])});
    $('date-error').textContent='';$('date-save-status').textContent='';
    dialog=modals.open({title:'Adjust Date / Time',description:'Time zone: '+current.zone,size:'medium',content:$('date-editor'),dismissOnBackdrop:false,onClose:function(){dialog=null},actions:[{label:'Cancel',id:'date-cancel',run:function(handle){handle.close()}},{label:'Save',id:'date-save',primary:true,run:save}]});
  }
  function save(handle){
    var values=[];
    for(var i=0;i<fields.length;i++){var text=$('date-'+fields[i][0]).textContent.trim();var n=Number(text);if(!/^\d+$/.test(text)||n<fields[i][2]||n>fields[i][3]){$('date-error').textContent='Enter a valid '+fields[i][1].toLowerCase()+'.';return;}values.push(n);}
    var days=new Date(Date.UTC(values[0],values[1],0)).getUTCDate();
    if(values[2]>days){$('date-error').textContent='That day does not exist in the selected month.';return;}
    var packed=String(values[0])+values.slice(1).map(pad).join('');
    if(App.command('clock.set',Number(packed))<0){$('date-error').textContent='Unable to start saving. Please try again.';return;}
    handle.setBusy(true);$('date-save-status').textContent='Saving…';
  }
  function saved(error){if(!dialog)return;dialog.setBusy(false);$('date-save-status').textContent='';if(error){$('date-error').textContent=error;return;}dialog.close();}
  popovers.attach($('clock'),$('calendar-popover'));$('adjust-datetime').onclick=open;
  window.dateTime={update:update,saved:saved,open:open};
})();
