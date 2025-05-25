module.exports = function (config) {
    config.set({
      basePath: '',
      frameworks: ['jasmine', '@angular-devkit/build-angular'],
      plugins: [
        require('karma-jasmine'),
        require('karma-chrome-launcher'),
        require('karma-jasmine-html-reporter'),
        require('karma-coverage'),
        require('@angular-devkit/build-angular/plugins/karma')
      ],
      client: {
        clearContext: false
      },
      reporters: ['progress', 'kjhtml', 'coverage'],
      coverageReporter: {
        dir: require('path').join(__dirname, './coverage'),
        subdir: 'competition-frontend',
        reporters: [
          { type: 'html' },
          { type: 'lcovonly' },
          { type: 'text-summary' }
        ]
      },
      browsers: ['ChromeHeadless'],
      restartOnFileChange: true
    });
  };