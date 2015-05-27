var db;
var dbname = 'nextone';
var storeTaskName = 'task';

var addtaskElt = document.querySelector('#addtask');
var taskslistElt = document.querySelector('#taskslist');


openTaskDb(function(query) {
  	taskslistElt.addEventListener('click', function(e) {
		if (e.target.classList.contains('delete')) {
		  	// Because the ID is stored in the DOM, it becomes
		  	// a string. So, we need to make it an integer again.
			query.delete(getTaskId(e.target), function() {
		    	// Refresh the to-do list
	    		renderAllTasks(query.list);
	  		});
		} else if (e.target.classList.contains('status')) {
			query.getAndModify(getTaskId(e.target), function(task, put) {
				switch (task.status) {
					case 'pending':
						task.status = 'done';
						break;
					case 'done':
						task.status = 'pending';
						break;
				}
				put(task);
			}, function() {
				renderAllTasks(query.list);
			});
		} else if (getTaskElement(e.target) !== null) {
			localStorage.setItem("currentTask", getTaskId(e.target))
			renderAllTasks(query.list);
		}	    	
	});
	document.body.addEventListener('submit', function onSubmit(e) {
		e.preventDefault();
		var task = {};
		task.text = addtaskElt.value.trim();
		task.status = 'pending';
		if (task.text !== "") {
			query.put(task, function() {
			  renderAllTasks(query.list);
			  addtaskElt.value = '';
			});
		}
	});
	renderAllTasks(query.list);
});

function getTaskId(e) {
	var taskElement = getTaskElement(e);
	if (taskElement) {
		return parseInt(taskElement.getAttribute('id'), 10);	
	}
	return -1;
}

function getTaskElement(e) {
	while (!e.classList.contains('task') || e === null) {
		e = e.parentElement;
	}
	return e;
}


function renderAllTasks(forEachTask) {
	taskslistElt.innerHTML = '';
	forEachTask(function(task) {	
		taskslistElt.innerHTML += renderTask(task, {});	
	});
}	

function renderTask(task, options) {
	var currentTaskId = localStorage.getItem("currentTask");
	var class_ = '';
	if (currentTaskId === ("" +task.timeStamp)) {
		class_ = 'current';
	}
	var html = '';

	html += '<li class="task '+ class_ +'"  id="'+task.timeStamp+'">';
	// Display date
	html += '<span class="date">';
	html += '<span class="day">';
	html += new Date(task.timeStamp).toLocaleDateString();
	html += '</span></span>';
	// Display status
	html += '<span class="status '+ task.status +'">';
	html += task.status;
	html += '</span>';
	// Display text
	html += '<span class="text">';
	html += task.text;
	html += '</span>';
	
	if (options.canDelete) {
		html += '<img class="delete" src="delete.svg">';
	}
	html += '</li>';
	return html;
}



function openTaskDb(callback) {
  // Open a database, specify the name and version
  var version = 6;
  var request = indexedDB.open(dbname, version);
  // Run migrations if necessary
  request.onupgradeneeded = function(e) {
    db = e.target.result;
    var transaction = e.target.transaction;
    transaction.onerror = dbError;
    var storeTask;
    if (db.objectStoreNames.contains(storeTaskName)) {
      // migrate  task data
      storeTask = transaction.objectStore(storeTaskName);
      storeTask.openCursor().onsuccess = function(e) {
        var result = e.target.result;
        // If there's data, add it to array
        if (result) {
          var task = result.value;
          // in version 3 add status attribute to task store
          if (task.status === undefined) {
            task.status = "pending";  
          }
          storeTask.put(task);
          result.continue();
        // Reach the end of the data
        }
      };
      
    } else {
      storeTask = db.createObjectStore(storeTaskName, { keyPath: 'timeStamp' });
    }
    if(!storeTask.indexNames.contains('by_status')) {
      var statusIndex = storeTask.createIndex('by_status', 'status');
    }    
    
  };
  request.onsuccess = function(e) {
    db = e.target.result;
    callback(tasksFunctions);
  };
  request.onerror = dbError;
}

var tasksFunctions = {};

tasksFunctions.put = function (task, callback) {
  var transaction = db.transaction([storeTaskName], 'readwrite');
  var store = transaction.objectStore(storeTaskName);
  if (task.timeStamp === undefined) {
    task.timeStamp = Date.now();
  }

  var request = store.put(task);

  transaction.oncomplete = function(e) {
    callback();
  };
  request.onerror = dbError;
}

tasksFunctions.getAndModify = function (id, callback, donecallback) {
  var transaction = db.transaction([storeTaskName], 'readwrite');
  var store = transaction.objectStore(storeTaskName);

  var request = store.get(id);

  request.onsuccess = function(event) {
    var task = request.result;
    callback(task, function (task) {
      store.put(task);
    });
  };
  transaction.oncomplete = function(e) {
    donecallback();
  };
  request.onerror = dbError;
}

tasksFunctions.delete = function (id, callback) {
  var transaction = db.transaction([storeTaskName], 'readwrite');
  var store = transaction.objectStore(storeTaskName);
  var request = store.delete(id);
  transaction.oncomplete = function(e) {
    callback();
  };
  request.onerror = dbError;
} 

tasksFunctions.list = function (callback) {
  var transaction = db.transaction([storeTaskName], 'readonly');
  var store = transaction.objectStore(storeTaskName);
  var request = store.index('by_status').openCursor(null, 'prev');
  // Get everything in the store
  //var request = store.openCursor(null, 'next');

  request.onsuccess = function(e) {
    var result = e.target.result;

    // If there's data, add it to array
    if (result) {
      callback(result.value);
      result.continue();
    } 
  };
}

function dbError(e) {
  console.error('An IndexedDB error has occurred', e);
}
