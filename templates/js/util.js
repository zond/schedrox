String.prototype.format = function() {
	var args = arguments;
	return this.replace(/{(\d+)}/g, function(match, number) { 
		return typeof args[number] != 'undefined'
		? args[number]
		: match
		;
	});
};
function putBelow(e1, e2) {
	var pos = e1.offset();
	e2.css('top', '' + (pos.top + e1.height()) + 'px');
	e2.css('left', '' + pos.left + 'px');
};
eventParticipantsDefaultFormat = '{0}/{1}-{2} ({3}/{4})';
/** 
* Get the ISO week date week number 
*/  
Date.prototype.getWeek = function () {  
	// Create a copy of this date object  
	var target  = new Date(this.valueOf());  

	// ISO week date weeks start on monday  
	// so correct the day number  
	var dayNr   = (this.getDay() + 6) % 7;  

	// ISO 8601 states that week 1 is the week  
	// with the first thursday of that year.  
	// Set the target date to the thursday in the target week  
	target.setDate(target.getDate() - dayNr + 3);  

	// Store the millisecond value of the target date  
	var firstThursday = target.valueOf();  

	// Set the target to the first thursday of the year  
	// First set the target to january first  
	target.setMonth(0, 1);  
	// Not a thursday? Correct the date to the next thursday  
	if (target.getDay() != 4) {  
		target.setMonth(0, 1 + ((4 - target.getDay()) + 7) % 7);  
	}  

	// The weeknumber is the number of weeks between the   
	// first thursday of the year and the thursday in the target week  
	return 1 + Math.ceil((firstThursday - target) / 604800000); // 604800000 = 7 * 24 * 3600 * 1000  
}  
/** 
* Get the ISO week date year number 
*/  
Date.prototype.getWeekYear = function ()   
{  
	// Create a new date object for the thursday of this week  
	var target  = new Date(this.valueOf());  
	target.setDate(target.getDate() - ((this.getDay() + 6) % 7) + 3);  

	return target.getFullYear();  
}  
Date.prototype.oldToISOString = Date.prototype.toISOString;
Date.prototype.toISOString = function() {
	var pushed = new Date(this.getTime());
	pushed.setHours(pushed.getHours() - pushed.getTimezoneOffset() / 60);
	return pushed.oldToISOString();
};
Date.oldParse = Date.parse;
Date.prototype.getISOTime = function() {
  return this.getTime() - this.getTimezoneOffset() * 60 * 1000;
};
Date.fromISOTime = function(t) {
  var at = new Date(t);
  return new Date(at.getTime() + at.getTimezoneOffset() * 60 * 1000);
};
Date.parse = function(s) {
	var ts = this.oldParse(s);
	var at = new Date(ts);
	return ts + at.getTimezoneOffset() * 60 * 1000;
};
hoursMinutesForDates = function(d1, d2) {
	return hoursMinutesForMinutes((d2.getTime() - d1.getTime()) / (1000 * 60));
};
hoursMinutesForMinutes = function(minutes) {
	var hours = parseInt(minutes / 60);
	minutes = minutes - hours * 60;
	var mStr = '' + minutes;
	var hStr = '' + hours;
	while (mStr.length < 2) {
		mStr = '0' + mStr;
	}
	while (hStr.length < 2) {
		hStr = '0' + hStr;
	}
	return hStr + ':' + mStr;
};
String.prototype.htmlEscape = function() {
	return $("<div></div>").text(this).html();
};
String.prototype.isEmail = function() {
	return /^(([^<>()[\]\\.,;:\s@\"]+(\.[^<>()[\]\\.,;:\s@\"]+)*)|(\".+\"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/.test(this);
};
String.prototype.hash = function() {
	var shaObj = new jsSHA(this, "TEXT");
	return shaObj.getHash("SHA-1", "HEX");
};
afterDayTime = function(a, b) {
	return a.getHours() > b.getHours() || (a.getHours() == b.getHours() && a.getMinutes() > b.getMinutes()) || (a.getHours() == b.getHours() && a.getMinutes() == b.getMinutes() && a.getSeconds() > b.getSeconds());
};
beforeDayTime = function(a, b) {
	return a.getHours() < b.getHours() || (a.getHours() == b.getHours() && a.getMinutes() < b.getMinutes()) || (a.getHours() == b.getHours() && a.getMinutes() == b.getMinutes() && a.getSeconds() < b.getSeconds());
};
betweenDayTimes = function(start, end, t) {
	return !beforeDayTime(t, start) && !afterDayTime(t, end);
};
luminosity = function(col) {
	return parseInt(Math.sqrt(
		col[0] * col[0] * .241 + 
			col[1] * col[1] * .691 + 
	col[2] * col[2] * .068));
};
parseColor = function(color) {

	var cache, p = parseInt, color = color.replace(/\s\s*/g,'');

	if (cache = /^#([\da-fA-F]{2})([\da-fA-F]{2})([\da-fA-F]{2})/.exec(color)) {
		cache = [p(cache[1], 16), p(cache[2], 16), p(cache[3], 16)];
	} else if (cache = /^#([\da-fA-F])([\da-fA-F])([\da-fA-F])/.exec(color)) {
		cache = [p(cache[1], 16) * 17, p(cache[2], 16) * 17, p(cache[3], 16) * 17];
	} else if (cache = /^rgba\(([\d]+),([\d]+),([\d]+),([\d]+|[\d]*.[\d]+)\)/.exec(color)) {
		cache = [+cache[1], +cache[2], +cache[3], +cache[4]];
	} else if (cache = /^rgb\(([\d]+),([\d]+),([\d]+)\)/.exec(color)) {
		cache = [+cache[1], +cache[2], +cache[3]];
	} else {
		throw Error(color + ' is not supported by $.parseColor'); 
	}
	isNaN(cache[3]) && (cache[3] = 1);
	return cache.slice(0,3 + !!$.support.rgba);
};
cbCounter = function(n, cb) {
	var that = this;
	that.n = n;
	this.call = function() {
		that.n--;
		if (that.n == 0) {
			if (cb != null) {
				cb();
			}
		}
	}
};
overlaps = function(a1, a2, b1, b2) {
	a1 = a1.getTime();
	a2 = a2.getTime();
	b1 = b1.getTime();
	b2 = b2.getTime()
	return (a1 >= b1 && a1 < b2) || (a2 > b1 && a2 <= b2) || (a1 <= b1 && a2 >= b2);
};
anyDayTimeConverterNoSeconds = new AnyTime.Converter();
anyDayTimeConverterNoSeconds.fmt = '{{.I "any_day_time_format_no_seconds" }}';
anyDateConverter = new AnyTime.Converter();
anyDateConverter.fmt = '{{.I "any_date_format" }}';
anyDayTimeConverter = new AnyTime.Converter();
anyDayTimeConverter.fmt = '{{.I "any_day_time_format" }}';
anyTimeConverter = new AnyTime.Converter();
anyTimeConverter.fmt = '{{.I "any_time_format" }}';
anyTimeConverterNoSeconds = {
	format: function(t) {
			return anyDateConverter.format(t) + " " + anyDayTimeConverterNoSeconds.format(t);
	}
};
gravatarImage = function(hash, options) {
	if (hash == null || hash == "") {
		hash = "00000000000000000000000000000000";
	}
	var params = [];
	for (var name in options) {
		params.push(encodeURIComponent(name) + '=' + encodeURIComponent(options[name]));
	}
	if (params.length > 0) {
		return 'https://www.gravatar.com/avatar/' + hash + '?d=retro&' + params.join('&');
	} else {
		return 'https://www.gravatar.com/avatar/' + hash + '?d=retro';
	}
};
isAuthorizedAny = function(auths, isClosed, isOwner, isAdmin, match) {
	if (isAdmin) {
		return true;
	}
	if (isClosed) {
	  return false;
	}
	if (isOwner) {
		return true;
	}
	if (app.getDomain() == null) {
		return false;
	}
	if (auths != null) {
		return _.any(auths, function(auth) {
			var auth_type = app.authTypes[auth.auth_type];
			if (auth.auth_type == match.auth_type) {
				if (!auth_type.has_write || auth.write || !match.write) {
					if (!auth_type.has_location || auth.location == null || match.location == null || auth.location == match.location) {
						if (!auth_type.has_event_kind || auth.event_kind == null || match.event_kind == null || auth.event_kind == match.event_kind) {
							if (!auth_type.has_event_type || auth.event_type == null || match.event_type == null || auth.event_type == match.event_type) {
								if (!auth_type.has_participant_type || auth.participant_type == null || match.participant_type == null || auth.participant_type == match.participant_type) {
									if (!auth_type.has_role || (auth.role == '' || auth.role == null) || (match.role == '' || match.role == null) || auth.role == match.role) {
										return true;
									}
								}
							}
						}
					}
				}
			}
			return false;
		});
	}
	return false
};
isAuthorized = function(auths, isClosed, isOwner, isAdmin, match) {
	if (isAdmin) {
		return true;
	}
	if (isClosed) {
	  return false;
	}
	if (isOwner) {
		return true;
	}
	if (app.getDomain() == null) {
		return false;
	}
	if (auths != null) {
		return _.any(auths, function(auth) {
			var auth_type = app.authTypes[auth.auth_type];
			if (auth.auth_type == match.auth_type) {
				if (!auth_type.has_write || auth.write || !match.write) {
					if (!auth_type.has_location || auth.location == null || auth.location == match.location) {
						if (!auth_type.has_event_kind || auth.event_kind == null || auth.event_kind == match.event_kind) {
							if (!auth_type.has_event_type || auth.event_type == null || auth.event_type == match.event_type) {
								if (!auth_type.has_participant_type || auth.participant_type == null || auth.participant_type == match.participant_type) {
									if (!auth_type.has_role || (auth.role == '' || auth.role == null) || auth.role == match.role) {
										return true;
									}
								}
							}
						}
					}
				}
			}
			return false;
		});
	}
	return false
};
jQuery.timeago.settings.strings = {
	prefixAgo: {{.I "timeago_prefixAgo" }},
	prefixFromNow: {{.I "timeago_prefixFromNow" }},
	suffixAgo: {{.I "timeago_suffixAgo" }},
	suffixFromNow: {{.I "timeago_suffixFromNow" }},
	seconds: {{.I "timeago_seconds" }},
	minute: {{.I "timeago_minute" }},
	minutes: {{.I "timeago_minutes" }},
	hour: {{.I "timeago_hour" }},
	hours: {{.I "timeago_hours"}},
	day: {{.I "timeago_day" }},
	days: {{.I "timeago_days" }},
	month: {{.I "timeago_month" }},
	months: {{.I "timeago_months" }},
	year: {{.I "timeago_year" }},
	years: {{.I "timeago_years" }},
	wordSeparator: {{.I "timeago_wordSeparator" }},
	numbers: {{.I "timeago_numbers" }},
};

salary_property_types = [
	'free_text',
	'integer',
	'float',
	'select',
];

today = function() {
	var d = new Date();
	return new Date(d.getFullYear(), d.getMonth(), d.getDate());
};

tomorrow = function() {
	var d = new Date();
	return new Date(d.getFullYear(), d.getMonth(), d.getDate() + 1);
};

firstSalaryBreakpointAfter = function(d) {
	var forwardOne = new Date(d.getFullYear(), d.getMonth(), d.getDate() + 1)
	var forwardOneNumber = 0;
	var nowNumber = 0;
	var breakpoint;
	if (app.getDomain().get('salary_config').salary_period == 'weekly') {
		forwardOneNumber = forwardOne.getDay();
		nowNumber = d.getDay();
		breakpoint = app.getDomain().get('salary_config').salary_breakpoint || 0;
	} else {
		forwardOneNumber = forwardOne.getDate()
		nowNumber = d.getDate();
		breakpoint = app.getDomain().get('salary_config').salary_breakpoint || 1;
	}
	if (forwardOneNumber == breakpoint || (forwardOneNumber > breakpoint && forwardOneNumber < nowNumber)) {
		return new Date(forwardOne.getFullYear(), forwardOne.getMonth(), forwardOne.getDate());
	} else {
		return firstSalaryBreakpointAfter(forwardOne);
	}
};

lastSalaryBreakpointBefore = function(d) {
	var backOne = new Date(d.getFullYear(), d.getMonth(), d.getDate() - 1)
	var backOneNumber = 0;
	var nowNumber = 0;
	var breakpoint;
	if (app.getDomain().get('salary_config').salary_period == 'weekly') {
		backOneNumber = backOne.getDay();
		nowNumber = d.getDay();
		breakpoint = app.getDomain().get('salary_config').salary_breakpoint || 0;
	} else {
		backOneNumber = backOne.getDate()
		nowNumber = d.getDate();
		breakpoint = app.getDomain().get('salary_config').salary_breakpoint || 1;
	}
	if (backOneNumber == breakpoint || (backOneNumber < breakpoint && backOneNumber > nowNumber)) {
		return new Date(backOne.getFullYear(), backOne.getMonth(), backOne.getDate());
	} else {
		return lastSalaryBreakpointBefore(backOne);
	}
};

getReportMimeType = function() {
	return("text/html; charset=UTF-8");
};
getReportContent = function() {
	return("<p>{{.I "To make the salary system work, you must add two JavaScript functions to the salary code: <i>getReportMimeType(reportData)</i> and <i>getReportContent(reportData)</i>."}}</p>");
};
