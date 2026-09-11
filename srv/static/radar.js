// radar.js — mobile collapse toggles + client-side filter chips. No deps.
(function(){
  var cards=[].slice.call(document.querySelectorAll('.list .card'));
  // collapse toggles
  cards.forEach(function(c){
    var t=c.querySelector('.tog');if(!t)return;
    t.addEventListener('click',function(){var o=c.classList.toggle('open');t.setAttribute('aria-expanded',o);});
  });
  // deep link (#fNN) opens that card
  if(location.hash){var h=document.getElementById(location.hash.slice(1));if(h){h.classList.add('open');var d=h.querySelector('details');if(d)d.open=true;h.scrollIntoView({block:'center'});}}
  // filters
  var bar=document.getElementById('filters'),count=document.getElementById('count');if(!bar)return;
  var bs=[].slice.call(bar.querySelectorAll('button'));
  function match(c,f){
    var d=c.dataset;
    switch(f){
      case '':return true;
      case 'new':return d.new==='1';
      case 'fresh':return d.fresh==='1';
      case 'unseen':return d.seen!=='1'||d.seenNow==='1'; // cards seen this visit stay until the filter is re-applied (Feedly-style sweep)
      case 'top':return +d.score>=70;
      case 'deadline':return !!d.deadline;
      case 'hard':return d.verdict==='hard to fill';
      case 'live':return d.verdict!=='closed'&&d.verdict!=='gone';
      case 'up':return d.vote==='1';
      case 'pinned':return d.pinned==='1';
      case 'open':return d.status==='open';
      case 'soon':return d.soon==='1'&&d.status!=='skip'&&d.status!=='rejected';
      case 'ssa':case 'at':case 'eu':return (d.region||d.track)===f;
      default:return d.track===f||d.kind===f||d.region===f;
    }
  }
  // venture segment (funding radar): orthogonal to the chips; sticky in URL (?p=) and localStorage
  var seg=document.getElementById('proj'),ps=seg?[].slice.call(seg.querySelectorAll('button')):[],proj='';
  function projOk(el){if(!seg||!proj)return true;var v=el.dataset.proj;return v===proj||v==='both';}
  function setProj(p,quiet){
    proj=p||'';
    ps.forEach(function(b){b.setAttribute('aria-selected',b.dataset.p===proj?'true':'false');});
    try{localStorage.setItem('funding.proj',proj);}catch(e){}
    var ups=[].slice.call(document.querySelectorAll('.upc')),nd=0;
    ups.forEach(function(u){var m=projOk(u);u.classList.toggle('off',!m);if(m)nd++;});
    var dh=document.getElementById('due-h');if(dh)dh.classList.toggle('off',!nd);
    ['n-due','n-due2'].forEach(function(id){var el=document.getElementById(id);if(el)el.textContent=nd;});
    var ne=document.getElementById('n-entries'),na=document.getElementById('n-all');
    var vis=cards.filter(function(c){return projOk(c)&&(showHidden||c.dataset.hidden!=='1');}).length;
    if(ne)ne.textContent=vis;if(na)na.textContent=vis;
    var un=document.getElementById('unseen-n'),us=document.getElementById('unseen-stat');
    if(un&&us&&!us.classList.contains('done')){var nu=cards.filter(function(c){return projOk(c)&&c.dataset.hidden!=='1'&&c.dataset.seen!=='1'&&c.dataset.seenNow!=='1';}).length;un.textContent=nu;us.hidden=!nu;}
    if(!quiet)document.documentElement.classList.toggle('proj-on',!!proj);
  }
  if(seg){
    ps.forEach(function(b){var n=cards.filter(function(c){var v=c.dataset.proj;return (!b.dataset.p||v===b.dataset.p||v==='both')&&c.dataset.hidden!=='1';}).length;var el=b.querySelector('.n');if(el)el.textContent=n;});
    ps.forEach(function(b){b.addEventListener('click',function(){if(b.dataset.p===proj)return;sweep();setProj(b.dataset.p);apply(cur());});});
    seg.addEventListener('keydown',function(e){var i=ps.indexOf(document.activeElement);if(i<0)return;var j=e.key==='ArrowRight'?i+1:e.key==='ArrowLeft'?i-1:-1;if(j<0||j>=ps.length)return;e.preventDefault();ps[j].focus();ps[j].click();});
  }
  function cur(){var b=bar.querySelector('button[aria-pressed=true]');return b?b.dataset.f:'';}
  var showHidden=/[?&]hidden=1/.test(location.search);
  var list=document.querySelector('.list'),emptyEl;
  function empty(f,n){
    if(emptyEl){emptyEl.remove();emptyEl=null;}
    if(n||!f||!list)return;
    var caught=f==='unseen'||f==='new'||f==='fresh';
    var lbl=(bs.filter(function(b){return b.dataset.f===f;})[0]||{}).textContent||f;
    emptyEl=document.createElement('div');emptyEl.className='empty'+(caught?'':' neutral');
    emptyEl.innerHTML='<div class="ring"><svg viewBox="0 0 24 24">'+(caught?'<path d="M20 6 9 17l-5-5"/>':'<circle cx="11" cy="11" r="7"/><path d="m20 20-3.5-3.5"/>')+'</svg></div>'
      +'<h3>'+(caught?'You\u2019re all caught up':'Nothing matches \u201c'+lbl+'\u201d')+'</h3>'
      +'<p>'+(caught?'Every item here has crossed your screen. New ones will show up after the next fetch.':'Try another filter or clear this one.')+'</p>'
      +'<button type="button">Show everything</button>';
    emptyEl.querySelector('button').addEventListener('click',function(){sweep();apply('');});
    list.appendChild(emptyEl);
  }
  function apply(f){
    var n=0;cards.forEach(function(c){var m=projOk(c)&&match(c,f)&&(showHidden||c.dataset.hidden!=='1');c.hidden=!m;if(m)n++;});
    empty(f,n);
    bs.forEach(function(b){var on=b.dataset.f===f;b.setAttribute('aria-pressed',on?'true':'false');if(on&&b.scrollIntoView&&bar.scrollWidth>bar.clientWidth)b.scrollIntoView({block:'nearest',inline:'center'});});
    [].forEach.call(document.querySelectorAll('[data-filter]'),function(a){a.classList.toggle('on',a.dataset.filter===f);});
    var us=document.getElementById('unseen-stat');if(us&&us.classList.contains('done')&&f!=='unseen'&&!us.classList.contains('bye')){us.classList.add('bye');setTimeout(function(){us.hidden=true;},520);var ub=bar.querySelector('button[data-f=unseen]');if(ub)ub.hidden=true;}
    if(count)count.textContent=n+'/'+cards.filter(function(c){return projOk(c)&&(showHidden||c.dataset.hidden!=='1');}).length;
    try{var hq=showHidden?'hidden=1':'',pq=proj?'p='+proj:'';history.replaceState(null,'',location.pathname+(f||hq||pq?'?'+[hq,pq,f?'f='+f:''].filter(Boolean).join('&'):'')+location.hash);}catch(e){}
  }
  function sweep(){cards.forEach(function(c){delete c.dataset.seenNow;});}
  bs.forEach(function(b){b.addEventListener('click',function(){sweep();apply(b.dataset.f);});});
  // stat chips elsewhere on the page (e.g. “5 new” in the header) toggle the same filter
  [].forEach.call(document.querySelectorAll('[data-filter]'),function(a){a.addEventListener('click',function(e){
    e.preventDefault();var f=a.dataset.filter,cur=bar.querySelector('button[aria-pressed=true]');sweep();
    apply(cur&&cur.dataset.f===f?'':f);
    var h2=bar.previousElementSibling;(h2||bar).scrollIntoView({behavior:'smooth',block:'start'});
  });});
  var sp=new URLSearchParams(location.search),q=sp.get('f');
  if(seg){var p0=sp.has('p')?sp.get('p'):(function(){try{return localStorage.getItem('funding.proj')||'';}catch(e){return '';}})();setProj(ps.some(function(b){return b.dataset.p===p0;})?p0:'');}
  apply(q&&bs.some(function(b){return b.dataset.f===q;})?q:'');
  window.radarRefilter=function(){if(seg)setProj(proj,true);apply(cur());};
})();

// seen tracking (Feedly/Slack-style): a card counts as seen once ≥50% of it has
// been on screen for ~1s. Marked cards lose their left marker, the header
// "N unseen" ticks down, and ids are POSTed in debounced batches (sendBeacon on
// pagehide so the last screen is not lost). Per user, server-side.
(function(){
  if(!('IntersectionObserver' in window))return;
  var list=document.querySelector('.list');if(!list)return;
  var radar=(document.querySelector('.react')||{dataset:{}}).dataset.radar||(location.pathname.indexOf('/funding')>-1?'grant':'job');
  var url='/admin/'+(radar==='grant'?'funding':'jobs')+'/seen';
  var cards=[].slice.call(list.querySelectorAll('.card[data-id]')).filter(function(c){return c.dataset.seen!=='1';});
  if(!cards.length)return;
  var nEl=document.getElementById('unseen-n'),stat=document.getElementById('unseen-stat');
  var queue=[],timers=new Map(),flushT;
  var reduce=matchMedia&&matchMedia('(prefers-reduced-motion:reduce)').matches;
  function tick(){
    if(!nEl)return;var n=Math.max(0,(+nEl.textContent||0)-1);
    nEl.textContent=n;nEl.classList.remove('roll');void nEl.offsetWidth;nEl.classList.add('roll');
    if(n||!stat||stat.classList.contains('done'))return;
    // 0 reached: swap the eye for a check, say "all caught up", hold, fold the pill away.
    var a=stat.querySelector('a.fl'),use=stat.querySelector('use'),lbl=stat.querySelector('.lbl');
    stat.classList.remove('hi');stat.classList.add('done');
    if(use&&document.getElementById('i-check'))use.setAttribute('href','#i-check');
    if(lbl)lbl.textContent='all caught up';
    if(a){a.title='Nothing unseen left';a.removeAttribute('href');}
    var on=a&&a.classList.contains('on'); // filter active: keep the pill while the user is in the unseen view
    setTimeout(function(){
      if(on)return;
      stat.classList.add('bye');
      var chip=document.querySelector('#filters button[data-f=unseen]');
      if(chip&&chip.getAttribute('aria-pressed')!=='true'){chip.classList.add('bye');setTimeout(function(){chip.hidden=true;},320);}
      setTimeout(function(){stat.hidden=true;},reduce?0:520);
    },reduce?1200:2600);
  }
  function flush(beacon){
    if(!queue.length)return;var ids=queue.splice(0),body=JSON.stringify({ids:ids});clearTimeout(flushT);
    if(beacon&&navigator.sendBeacon){try{if(navigator.sendBeacon(url,new Blob([body],{type:'application/json'})))return;}catch(e){}}
    fetch(url,{method:'POST',credentials:'same-origin',keepalive:true,headers:{'Content-Type':'application/json','Accept':'application/json'},body:body})
      .catch(function(){queue=ids.concat(queue);flushT=setTimeout(flush,10000);});
  }
  function mark(c){
    c.dataset.seen='1';c.dataset.seenNow='1';c.classList.add('seen-now');
    if(c.dataset.hidden!=='1')tick();
    queue.push(+c.dataset.id);clearTimeout(flushT);flushT=setTimeout(flush,queue.length>=25?200:2000);
  }
  var io=new IntersectionObserver(function(es){
    es.forEach(function(e){var c=e.target;
      // ≥50% of the card, or (for cards taller than half the screen) ≥50% of the viewport filled by it
      var vis=e.isIntersecting&&(e.intersectionRatio>=.5||e.intersectionRect.height>=innerHeight*.5);
      if(vis){if(!timers.has(c))timers.set(c,setTimeout(function(){timers.delete(c);io.unobserve(c);mark(c);},1000));}
      else if(timers.has(c)){clearTimeout(timers.get(c));timers.delete(c);}
    });
  },{threshold:[.1,.25,.5,.75]});
  cards.forEach(function(c){io.observe(c);});
  // tab hidden → the dwell timers must not fire blind
  document.addEventListener('visibilitychange',function(){if(document.hidden){timers.forEach(function(t){clearTimeout(t);});timers.clear();flush(true);}});
  window.addEventListener('pagehide',function(){flush(true);});
})();

// live "updating" indicator for /admin/jobs: poll status.json while a job runs, reload when done.
(function(){
  var st=document.getElementById('status');if(!st)return;
  var live=document.getElementById('live'),txt=document.getElementById('live-text'),upd=document.getElementById('upd');
  var running=st.dataset.running||'';
  function show(kind,since){
    running=kind;document.body.classList.toggle('busy',!!kind);
    live.hidden=!kind;upd.hidden=!!kind;
    if(kind){var s=since?Math.max(0,Math.round(Date.now()/1000-since)):0;txt.textContent=kind+'ing…'+(s>=5?' '+(s<60?s+'s':Math.round(s/60)+' min'):'');}
  }
  var wasRunning=!!running;
  show(running,+st.dataset.since||0);
  var delay=wasRunning?2000:15000,timer;
  function poll(){
    fetch('/admin/jobs/status.json',{credentials:'same-origin',cache:'no-store'}).then(function(r){return r.json();}).then(function(j){
      if(j.running){wasRunning=true;delay=2000;show(j.running,j.since);}
      else if(wasRunning){location.replace(location.pathname+(location.search||'')+location.hash);return;}
      else{delay=Math.min(delay*2,60000);}
      timer=setTimeout(poll,delay);
    }).catch(function(){timer=setTimeout(poll,10000);});
  }
  timer=setTimeout(poll,delay);
  document.addEventListener('visibilitychange',function(){if(!document.hidden){clearTimeout(timer);poll();}});
  // clicking a bar button: show indicator immediately (before redirect lands)
  var bar=document.getElementById('bar');if(bar)bar.addEventListener('submit',function(e){
    var a=(e.target.getAttribute('action')||'').split('/').pop();show(a,Math.round(Date.now()/1000));
  });
})();

// owner feedback: thumbs / note / status / trash without page reloads, plus a
// one-time "why?" prompt after trashing. Falls back to plain forms without JS.
(function(){
  document.documentElement.classList.add('js');
  var reacts=[].slice.call(document.querySelectorAll('.react'));if(!reacts.length)return;
  function post(url,data){
    var b=new URLSearchParams();Object.keys(data||{}).forEach(function(k){b.append(k,data[k]);});
    return fetch(url,{method:'POST',credentials:'same-origin',headers:{'Accept':'application/json','Content-Type':'application/x-www-form-urlencoded'},body:b.toString()})
      .then(function(r){if(!r.ok)throw new Error('HTTP '+r.status);return r.json();});
  }
  var filterCount=function(){var b=document.querySelector('#filters button[aria-pressed=true]');if(b)b.click();};
  reacts.forEach(function(R){
    var card=R.closest('.card'),radar=R.dataset.radar,id=R.dataset.id,saved=R.querySelector('.saved'),t;
    var base='/admin/'+(radar==='job'?'jobs':'funding')+'/';
    function flash(msg,err){saved.textContent=msg;saved.classList.toggle('err',!!err);saved.classList.add('show');clearTimeout(t);t=setTimeout(function(){saved.classList.remove('show');},err?4000:1500);}
    // thumbs
    var vb=[].slice.call(R.querySelectorAll('.vote .ib'));
    function pop(b){b.classList.remove('pop');void b.offsetWidth;b.classList.add('pop');}
    function paintVote(v){vb.forEach(function(b){var on=+b.value===v;b.classList.toggle('on',on);b.setAttribute('aria-pressed',on);});
      card.dataset.vote=v;card.classList.toggle('voted-up',v===1);card.classList.toggle('voted-down',v===-1);}
    vb.forEach(function(b){b.addEventListener('click',function(e){e.preventDefault();
      var v=+b.value,cur=+card.dataset.vote||0,nv=v===cur?0:v;paintVote(nv);if(nv)pop(b);
      post(base+'vote/'+id,{vote:v}).then(function(j){paintVote(j.vote);flash(j.vote?(j.vote>0?'more like this':'fewer like this'):'cleared');}).catch(function(){paintVote(cur);flash('failed',1);});
    });});
    // pin / star → moves the card to the top of the list
    var pb=R.querySelector('.pin .ib');if(pb){
      function paintPin(p){pb.classList.toggle('on',p);pb.setAttribute('aria-pressed',p);pb.dataset.tip=p?'Unpin':'Pin to top';card.dataset.pinned=p?'1':'';card.classList.toggle('pinned',p);}
      function place(p){var list=card.parentNode,cards=[].slice.call(list.querySelectorAll('.card'));
        var target=null;
        if(p){target=cards.find(function(c){return c!==card&&c.dataset.pinned!=='1';})||null;} // first unpinned card
        else{var lastPinned=null;cards.forEach(function(c){if(c!==card&&c.dataset.pinned==='1')lastPinned=c;});
          // back to score order among unpinned cards
          var sc=+card.dataset.score||-1;target=cards.find(function(c){return c!==card&&c.dataset.pinned!=='1'&&(+c.dataset.score||-1)<sc;})||null;
          if(lastPinned&&target&&lastPinned.compareDocumentPosition(target)&Node.PRECEDING_ORDER)target=lastPinned.nextElementSibling;}
        if(target===card.nextElementSibling||(target===null&&!card.nextElementSibling))return;
        var r0=card.getBoundingClientRect();card.classList.add('moving');
        setTimeout(function(){list.insertBefore(card,target);card.classList.remove('moving');
          if(p){var r1=card.getBoundingClientRect();if(r1.top<0||r1.bottom>innerHeight)card.scrollIntoView({block:'center',behavior:'smooth'});}
          else{window.scrollBy(0,card.getBoundingClientRect().top-r0.top);}},250);}
      pb.addEventListener('click',function(e){e.preventDefault();var cur=card.dataset.pinned==='1',nv=!cur;paintPin(nv);if(nv)pop(pb);
        post(base+'pin/'+id,{pinned:nv?'1':'0'}).then(function(j){paintPin(j.pinned);flash(j.pinned?'pinned to top':'unpinned');place(j.pinned);}).catch(function(){paintPin(cur);flash('failed',1);});});
    }
    // note
    var tog=R.querySelector('.note-tog'),nf=R.querySelector('.note-form'),ta=nf.querySelector('textarea'),last=ta.value,saveT;
    function openNote(o){nf.hidden=!o;tog.setAttribute('aria-expanded',o);if(o){ta.focus();ta.setSelectionRange(ta.value.length,ta.value.length);grow();}}
    function grow(){if(!('fieldSizing' in ta.style)){ta.style.height='auto';ta.style.height=Math.min(ta.scrollHeight+2,14*16)+'px';}}
    function saveNote(){var v=ta.value.trim();if(v===last){nf.classList.remove('dirty');return Promise.resolve();}
      return post(base+'note/'+id,{note:v}).then(function(j){last=j.note;nf.classList.remove('dirty');tog.classList.toggle('has',!!last);tog.dataset.tip=last?'Edit note':'Add note';flash('note saved');}).catch(function(){flash('not saved',1);});}
    tog.addEventListener('click',function(){openNote(nf.hidden);});
    ta.addEventListener('input',function(){grow();nf.classList.toggle('dirty',ta.value.trim()!==last);clearTimeout(saveT);saveT=setTimeout(saveNote,1200);});
    ta.addEventListener('blur',function(){clearTimeout(saveT);saveNote();});
    ta.addEventListener('keydown',function(e){
      if((e.metaKey||e.ctrlKey)&&e.key==='Enter'){e.preventDefault();clearTimeout(saveT);saveNote().then(function(){if(!ta.value.trim())openNote(false);});}
      else if(e.key==='Escape'){e.preventDefault();clearTimeout(saveT);saveNote();openNote(!!ta.value.trim()&&false);}
    });
    nf.addEventListener('submit',function(e){e.preventDefault();clearTimeout(saveT);saveNote();});
    // why? (asked once per item)
    var why=R.querySelector('.why-ask');
    function askWhy(undo){
      why.hidden=false;
      var chips=[].slice.call(why.querySelectorAll('.chips button'));
      chips.forEach(function(c){c.onclick=function(){chips.forEach(function(x){x.classList.toggle('on',x===c);});
        var data={reason:c.dataset.reason};if(radar==='grant')data.status=card.dataset.status||'skip';
        post(base+(radar==='job'?'hide/':'status/')+id,data).then(function(){card.dataset.asked='1';why.hidden=true;flash('thanks');}).catch(function(){flash('failed',1);});};});
      why.querySelector('.skip').onclick=function(){why.hidden=true;};
    }
    // trash / restore (jobs: hidden flag; grants: status skip/open)
    var tf=R.querySelector('.trash');if(tf){tf.addEventListener('submit',function(e){e.preventDefault();
      var un=!!(tf.querySelector('[name=unhide]')||(tf.querySelector('[name=status]')||{}).value==='open');card.classList.add('leaving');
      var req=radar==='job'?post(base+'hide/'+id,un?{unhide:'1'}:{}):post(base+'status/'+id,{status:un?'open':'skip'});
      req.then(function(j){
        if(radar==='grant'){card.dataset.status=j.status;card.classList.toggle('done',j.hidden);var tg=card.querySelector('.tags .tag.status');if(tg)tg.remove();
          if(j.status!=='open'){var sp=document.createElement('span');sp.className='tag status '+j.status;sp.textContent=j.status;card.querySelector('.tags').appendChild(sp);}}
        card.classList.remove('leaving');card.classList.toggle('hidden-row',j.hidden);card.dataset.hidden=j.hidden?'1':'';
        var showingHidden=/[?&]hidden=1/.test(location.search);
        var restore=radar==='job'?'<input type="hidden" name="unhide" value="1">':'<input type="hidden" name="status" value="open">';
        var skip=radar==='job'?'':'<input type="hidden" name="status" value="skip">';
        tf.innerHTML=j.hidden?restore+'<button class="ib" data-tip="Restore to list" aria-label="restore"><svg class="i"><use href="#i-undo"/></svg></button>'
          :skip+'<button class="ib" data-tip="Not relevant" aria-label="not relevant"><svg class="i"><use href="#i-trash"/></svg></button>';
        if(j.hidden&&j.ask_reason)askWhy();else why.hidden=true;
        if(j.hidden&&!showingHidden&&!j.ask_reason){setTimeout(function(){card.hidden=true;filterCount();},250);}
        else if(j.hidden&&!showingHidden){why.querySelector('.skip').addEventListener('click',function(){card.hidden=true;filterCount();},{once:true});
          [].forEach.call(why.querySelectorAll('.chips button'),function(c){c.addEventListener('click',function(){setTimeout(function(){card.hidden=true;filterCount();},600);},{once:true});});}
        flash(j.hidden?'removed':'restored');
      }).catch(function(){card.classList.remove('leaving');flash('failed',1);});
    });}

  });
})();
