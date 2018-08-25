L.MultiColorDivHeatmapLayer = L.FeatureGroup.extend({
	options: {
		gradient: {
			0: 'blue',
			0.3: 'cyan',
			0.5: 'lime',
			0.7: 'yellow',
			1: 'red'
		},
		max_value: 100,
		min_value: 0,
		radius: 70,
		use_gradient: true,
		clickable: false,// heatmap.on('click',function() {...});
		animation_delay: 20,
		default_opacity: 1,
		show_legend: true,
		legend_position: 'bottomleft',
	},

	initialize: function (options) {
		//set the options
		//L.Util.setOptions(this, options);
		this.setOptions(options);

		this._layers = {};
	},

	// adds a zone on map
	_addZone: function (lat, lng, value, opacity, old_marker) {
		// Remove previous
		if (typeof old_marker != 'undefined') {
			this.removeLayer(old_marker);
		}

		if ((!value && value != 0) || (!lat && lat != 0) || (!lng && lng != 0)) {
			throw new Error('Provide a latitude, longitude and a value');
		}

		if (opacity > 1 || opacity < 0) {
			// throw new Error('Opacity should be beetween 0 and 1');
			console.error("Opacity should be beetween 0 and 1");
			console.error("using default Opacity:" + this.options.default_opacity);
		}

		// Define the marker
		var alpha_start = opacity,
			alpha_end = !this.options.use_gradient ? opacity : 0,
			color = this._getGradientColorRGBA(value);

		var gradient = 'radial-gradient(circle closest-side, rgba(' + color.r + ', ' + color.g + ', ' + color.b + ', ' + alpha_start + ') 0%, rgba(' + color.r + ', ' + color.g + ', ' + color.b + ', ' + alpha_end + ') 100%)';
		var html = '<div data-value="' + value + '" style="width:100%;height:100%;border-radius:50%;background-image:' + gradient + '">';
		var size = this.options.radius;
		var divicon = L.divIcon({
			iconSize: [size, size],
			className: 'leaflet-heatmap-blob',
			html: html
		});

		var marker = L.marker([lat, lng], {
			icon: divicon,
			clickable: this.options.clickable,
			keyboard: false,
			opacity: opacity
		}).addTo(this);

		return marker;
	},

	_dataset: [],
	_markerset: [],

	// removes current data and add new one
	setData: function (data) {
		// Data object is three values [ {lat,lng,value}, {...}, ...]
		this.clearData();
		var self = this;
		data.forEach(function (point) {
			point.opacity = !point.opacity || point.opacity > 1 || point.opacity < 0 ? this.options.default_opacity : point.opacity;
			var marker = self._addZone(point.lat, point.lng, point.value, point.opacity);
			self._markerset.push(marker);
			self._dataset.push({
				lat: point.lat,
				lng: point.lng,
				value: point.value,
				opacity: point.opacity,
			});

		});
		if (this.options.show_legend && this._canvas) {
			this.legend();
		}
		return this;
	},

	// returns current data
	getData: function () {
		return this._dataset;
	},

	// remove all data
	clearData: function () {
		this.clearLayers();
		this._markerset = [];
		this._dataset = [];
	},

	// remove all layers and draw theme again
	reDraw: function () {
		this.clearLayers();
		this._markerset = [];
		var self = this;
		this._dataset.forEach(function (point) {
			var marker = self._addZone(point.lat, point.lng, point.value, point.opacity);
			self._markerset.push(marker);
		});
		if (this.options.show_legend && this._canvas) {
			this.legend();
		}
		return this;
	},

	// animates points on map by changing the opacity
	_animateZone: function (lat, lng, value, start_opacity, end_opacity, marker, fadeCallback) {
		var self = this;
		if (!marker) var marker;

		var o = start_opacity;
		var step = start_opacity < end_opacity ? 0.1 : -0.1;

		var seed = setInterval(function () {
			//if (!marker) self.removeLayer(marker);
			o = o + step;
			// Gate values so that the blob is always correct during progression
			o = o < 0 ? 0 : o > 1 ? 1 : o;
			marker = self._addZone(lat, lng, value, o, marker);
			if (o >= Math.max(start_opacity, end_opacity) || o <= Math.min(start_opacity, end_opacity)) {
				window.clearInterval(seed);
				if (fadeCallback) fadeCallback(marker);
			}
		}, this.options.animation_delay);

	},

	// replaces old data with new one with fade effect
	morphData: function (new_data) {
		this.fadeOutData();
		this.fadeInData(new_data);
		return this;
	},

	// fades in new data and add them to data set
	fadeInData: function (data) {
		var self = this;
		data.forEach(function (point) {
			point.opacity = !point.opacity || point.opacity > 1 || point.opacity < 0 ? this.options.default_opacity : point.opacity;
			self._animateZone(point.lat, point.lng, point.value, 0, point.opacity, null, function (marker) {
				self._markerset.push(marker);
				self._dataset.push({
					lat: point.lat,
					lng: point.lng,
					value: point.value,
					opacity: point.opacity,
				});
			});
		})
	},

	// removes old data whit a fade effect
	fadeOutData: function () {
		var self = this;
		var qty = self._dataset.length;
		for (var i = 0; i < qty; i++) {
			var point = self._dataset.pop();
			var marker = self._markerset.pop();
			self._animateZone(point.lat, point.lng, point.value, point.opacity, 0, marker, function (marker) {
				self.removeLayer(marker);
			});
		}
	},

	// this function generates a 256 color gradient and save it in _grad for later usage
	_generateGradient: function (grad) {
		// create a 256x1 gradient that we'll use to turn a grayscale heatmap into a colored one
		var canvas = this._createCanvas(),
			ctx = canvas.getContext('2d'),
			gradient = ctx.createLinearGradient(0, 0, 256, 0);

		canvas.width = 256;
		canvas.height = 30;

		for (var i in grad) {
			gradient.addColorStop(+i, grad[i]);
		}

		ctx.fillStyle = gradient;
		ctx.fillRect(0, 0, 256, 10);

		this._gradient = ctx.getImageData(0, 0, 256, 1).data;

		ctx.font = "20px Arial";
		ctx.fillStyle = 'black';
		// ctx.rotate(Math.PI / 2);
		ctx.fillText(this.options.min_value, 0, 30);
		ctx.fillText(this.options.max_value, 256 - ctx.measureText(this.options.max_value).width, 30);
		canvas.style.transform = "scale(0.5)";
		// document.body.insertBefore(self._canvas, document.body.firstChild);

		return canvas;
	},

	// generates a canvas element
	_createCanvas: function () {
		if (typeof document !== 'undefined') {
			return document.createElement('canvas');
		} else {
			// create a new canvas instance in node.js
			// the canvas class needs to have a default constructor without any parameter
			return new this._canvas.constructor();
		}
	},

	// fins equivalent color of a value in given range
	_getGradientColorRGBA: function (value) {
		start_index = Math.floor((value - this.options.min_value) * 256 / this._denominator) * 4; // 256 colors * 4 number for eache one (r, g, b, a)
		return {
			r: this._gradient[start_index],
			g: this._gradient[start_index + 1],
			b: this._gradient[start_index + 2],
			a: this._gradient[start_index + 3]
		}
	},

	// sets new max value
	setMaxValue: function (max) {
		this.setOptions({max_value: min});
		return this;
	},

	// sets new min value
	setMinValue: function (min) {
		this.setOptions({min_value: min});
		return this;
	},

	// changes value of given option name
	setOption(name, value) {
		o = {};
		o[name] = value;
		this.setOptions(o);
		return this;
	},

	// change options
	setOptions(options) {
		L.Util.setOptions(this, options);

		// calculates the value of denominator for color indexing
		this._denominator = this.options.max_value - this.options.min_value;

		// generate the gradient
		this._canvas = this._generateGradient(this.options.gradient);

		if (this.options.show_legend && this._canvas) {
			this.legend();
		}
		return this;
	},

	legend() {
		self = this;

		if (!this._legend) {
			this._legend = L.control();
			this._legend.options.position = this.options.legend_position;

			this._legend.onAdd = function (map) {
				if (!this._div) {
					this._div = L.DomUtil.create('div');
					this._div.className = "leaflet-control-attribution";
					this._div.appendChild(self._canvas);
				}
				return this._div;
			};

			if (this._map)
				this._legend.addTo(this._map);

		} else {
			this._legend.remove();
			this._legend = null;
			this.legend();
		}

		return this;
	},
});