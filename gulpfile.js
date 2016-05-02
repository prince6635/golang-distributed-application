var gulp = require('gulp');
var path = require('path');
var shell = require('gulp-shell');

var testGoPath = 'src/test/**/*.go';
var powerplantGoPath = 'src/powerplant/**/*.go'

gulp.task('compilepkg', function() {
  return gulp.src([testGoPath, powerplantGoPath], {read: false})
    .pipe(shell(['go install github.com/golang-distributed-application/src/<%= stripPath(file.path) %>'],
      {
          templateData: {
            stripPath: function(filePath) {
              var subPath = filePath.substring(process.cwd().length + 5);
              var pkg = subPath.substring(0, subPath.lastIndexOf(path.sep));
              return pkg;
            }
          }
      })
    );
});

gulp.task('watch', function() {
  gulp.watch([testGoPath, powerplantGoPath], ['compilepkg']);
});
