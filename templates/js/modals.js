myalert = function(text, cb) {
	$.modal("<div class=\"alert-content\"><p>" + text + "</p></div>", { 
		closeHTML: '<a href="#">×</a>',
		closeClass: 'absolute-close',
		overlayClose: true,
		minWidth: '350',
		autoResize: true,
		onClose: function() {
			$.modal.close();
			if (cb) {
				cb();
			}
		},
		onShow: function(dialog) {
			$('.simplemodal-wrap').css('overflow', 'auto');
			$('#simplemodal-container').css('padding-bottom', '200');
		},
	});
};
myconfirm = function(text, callback) {
	var innercb = null;
	var enterl = null;
	var close = null;
	innercb = function(ev) {
		ev.preventDefault();
		close();
		callback();
	};
	enterl = function(ev) {
		if (ev.which == 13) {
			innercb(ev);
		}
	};
	close = function() {
		$(document).unbind('keypress', enterl);
		$.modal.close();
	};
	$.modal("<div class=\"alert-content\"><p>" + text + "</p></div>", { 
		closeHTML: '<a href="#">×</a>',
		closeClass: 'absolute-close',
		overlayClose: true,
		minWidth: '30%',
		autoResize: true,
		onClose: close,
		onShow: function(dialog) {
			$('.simplemodal-wrap').css('overflow', 'visible');
			$(document).bind('keypress', enterl);
			var callback_container = $('<div class="simplemodal-callback-container"></div>');
			var callback_link = $('<a href="#" class="btn btn-primary simplemodal-callback">{{.I "Yes" }}</a>');
			callback_link.click(innercb);
			callback_container.append(callback_link);
			$('.simplemodal-wrap').append(callback_container);
		}
	});
};
var mymodalNiceclose = false;
mymodalClose = function() {
  mymodalNiceclose = true;
	$.modal.close();
};
mymodal = function(html, callbacks, options) {
	var closeCallback = null;
	var cancelCallback = null;
	if (callbacks != null) {
		closeCallback = callbacks.onClose;
		delete(callbacks['onClose']);
		cancelCallback = callbacks.onCancel;
		delete(callbacks['onCancel']);
	}
	var cancelled = true;
	$.modal(html, { 
		closeHTML: '<a href="#">×</a>',
		closeClass: 'absolute-close',
		overlayClose: true,
		minWidth: options && options.min_width ? options.min_width : '60%',
		minHeight: options && options.min_height ? options.min_height : '60%',
		maxHeight: options && options.max_height ? options.max_height : '80%',
		autoResize: true,
		onClose: function() {
			$.modal.close();
			if (closeCallback != null) {
				closeCallback();
			}
			if (cancelCallback != null && cancelled && !mymodalNiceclose) {
				cancelCallback();
			}
			mymodalNiceclose = false;
		},
		onShow: function(dialog) {
			$('.simplemodal-wrap').css('overflow', 'auto');
			var has_callbacks = false;
			for (var name in callbacks) {
				has_callbacks = true;
			}
			if (has_callbacks) {
				var callback_container = $('<div class="simplemodal-callback-container"></div>');
				for (var name in callbacks) {
					var callback_link = $('<a href="#" class="btn btn-primary simplemodal-callback">' + name + '</a>');
					callback_link.click(function(name) {
						return function(ev) {
							cancelled = false;
							ev.preventDefault();
							$.modal.close();
							callbacks[name]();
						}
					}(name));
					callback_container.append(callback_link);
				}
				$('.simplemodal-wrap').append(callback_container);
			}
		},
	});
};
