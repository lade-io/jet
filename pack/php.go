package pack

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/ake-persson/mapslice-json"
)

type PhpPack struct {
	WorkDir string
}

func (p *PhpPack) Detect() bool {
	return fileExists(p.WorkDir, "composer.json") ||
		fileExists(p.WorkDir, "index.php")
}

func (p *PhpPack) Metadata() *Metadata {
	meta := &Metadata{
		Path:    "/var/www",
		User:    "www-data",
		Variant: "apache",
	}
	meta.Depends = append(meta.Depends, &Depend{
		Args: []string{
			"ln -s php.ini-production $PHP_INI_DIR/php.ini",
			"a2enmod rewrite && chown www-data:www-data /var/www",
			"echo 'ServerName localhost' >> /etc/apache2/apache2.conf",
			"echo 'DocumentRoot ${APACHE_ROOT}' >> /etc/apache2/apache2.conf",
			"echo ': ${PORT:=3000}\\nexport PORT' >> /etc/apache2/envvars",
			"sed -i 's/^Listen.*/Listen ${PORT}/' /etc/apache2/ports.conf",
		},
	})
	meta.Env = map[string]string{
		"APACHE_ROOT": filepath.Join(meta.Path, p.webroot()) + "/",
	}

	if fileExists(p.WorkDir, "composer.json") {
		conf, core, load, pecl, pkgs := p.extensions()
		if len(conf) > 0 {
			meta.Depends = append(meta.Depends, &Depend{
				Name: "docker-php-ext-configure",
				Args: conf,
			})
		}

		if len(core) > 0 {
			meta.Depends = append(meta.Depends, &Depend{
				Name: "docker-php-ext-install",
				Args: core,
				List: true,
			})
		}

		if len(pecl) > 0 {
			meta.Depends = append(meta.Depends, &Depend{
				Name: "pecl install",
				Args: pecl,
				List: true,
			})
			meta.Depends = append(meta.Depends, &Depend{
				Name: "docker-php-ext-enable",
				Args: load,
				List: true,
			})
		}

		pkgs = append(pkgs, []string{"git", "wget", "zip"}...)
		meta.Packages = append(meta.Packages, pkgs...)
		meta.Tools = append(meta.Tools, &Tool{
			Name:    "composer",
			Owner:   "composer",
			Files:   []string{"composer.json", "composer.lock"},
			Install: []string{"install --no-dev --no-scripts"},
		})
	}
	return meta
}

func (p *PhpPack) Name() string {
	return "php"
}

func (p *PhpPack) Command() (string, error) {
	return "", nil
}

func (p *PhpPack) Version() (string, error) {
	requires := p.requires()
	version := phpPipe.Split(requires["php"], -1)
	return strings.Join(version, "||"), nil
}

func (p *PhpPack) extensions() ([]string, []string, []string, []string, []string) {
	requires := p.requires()
	var exts []string
	for require := range requires {
		ext := strings.TrimPrefix(require, "ext-")
		if ext != require {
			exts = append(exts, ext)
		}
	}
	sort.Strings(exts)

	var conf, core, load, pecl, pkgs []string
	seen := map[string]bool{}
	for _, ext := range exts {
		var args []string
		var ok bool
		if args, ok = phpCoreExts[ext]; ok {
			core = append(core, ext)
			if len(args) > 1 {
				ext += " " + args[1]
				conf = append(conf, ext)
			}
		} else if args, ok = phpPeclExts[ext]; ok {
			load = append(load, ext)
			if len(args) > 1 {
				ext += "-" + args[1]
			}
			pecl = append(pecl, ext)
		}
		if len(args) > 0 {
			args = strings.Split(args[0], ",")
		}
		for _, pkg := range args {
			if !seen[pkg] {
				pkgs = append(pkgs, pkg)
				seen[pkg] = true
			}
		}
	}
	sort.Strings(pkgs)
	return conf, core, load, pecl, pkgs
}

type composerLock struct {
	Packages []composerPackage `json:"packages"`
}

type composerPackage struct {
	Name    string            `json:"name"`
	Require mapslice.MapSlice `json:"require"`
}

func composerRequire(name string, pkgs map[string]mapslice.MapSlice, requires map[string]string) {
	for _, item := range pkgs[name] {
		key, ok := item.Key.(string)
		if !ok {
			continue
		}
		value, ok := item.Value.(string)
		if !ok {
			continue
		}
		if version, ok := requires[key]; !ok {
			requires[key] = value
			composerRequire(key, pkgs, requires)
		} else if !strings.Contains(version, value) {
			requires[key] = version + "," + value
		}
	}
}

func (p *PhpPack) requires() map[string]string {
	pkgs := map[string]mapslice.MapSlice{}
	b, err := fileRead(p.WorkDir, "composer.lock")
	if err == nil {
		lock := composerLock{}
		if err = json.Unmarshal(b, &lock); err == nil {
			for _, pkg := range lock.Packages {
				pkgs[pkg.Name] = pkg.Require
			}
		}
	}

	b, err = fileRead(p.WorkDir, "composer.json")
	if err != nil {
		return nil
	}

	conf := composerPackage{}
	if err = json.Unmarshal(b, &conf); err != nil {
		return nil
	}

	requires := map[string]string{}
	for _, item := range conf.Require {
		key, ok := item.Key.(string)
		if !ok {
			continue
		}
		value, ok := item.Value.(string)
		if !ok {
			continue
		}
		requires[key] = value
		composerRequire(key, pkgs, requires)
	}
	return requires
}

func (p *PhpPack) webroot() string {
	paths, err := fileGlob(p.WorkDir, "**/index.php")
	if err != nil || len(paths) == 0 {
		return ""
	}

	if len(paths) > 0 {
		sort.SliceStable(paths, func(i, j int) bool {
			return strings.Count(paths[i], "/") < strings.Count(paths[j], "/")
		})
		for _, path := range paths {
			file, err := os.Open(filepath.Join(p.WorkDir, path))
			if err != nil {
				continue
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			web := true
			for scanner.Scan() {
				line := scanner.Text()
				if phpLayout.MatchString(line) {
					web = false
					break
				}
			}
			if web {
				return filepath.Dir(path)
			}
		}
	}
	return filepath.Dir(paths[0])
}

var (
	phpLayout = regexp.MustCompile(`^require.*index\.php`)
	phpPipe   = regexp.MustCompile(`\|+`)
)

const (
	phpImap = "--with-kerberos --with-imap-ssl"
	phpODBC = "--with-pdo-odbc=unixODBC,/usr"
)

var phpCoreExts = map[string][]string{
	"bz2":          []string{"libbz2-dev"},
	"curl":         []string{"libcurl4-openssl-dev"},
	"dba":          []string{},
	"enchant":      []string{"libenchant-*dev"},
	"exif":         []string{},
	"fileinfo":     []string{},
	"ftp":          []string{"libssl-dev"},
	"gd":           []string{"libpng-dev"},
	"gettext":      []string{},
	"gmp":          []string{"libgmp-dev"},
	"imap":         []string{"libc-client-dev,libkrb5-dev", phpImap},
	"intl":         []string{"libicu-dev"},
	"ldap":         []string{"libldap2-dev"},
	"mbstring":     []string{"libonig-dev"},
	"mysqli":       []string{},
	"pcntl":        []string{},
	"pdo":          []string{},
	"pdo_firebird": []string{"firebird-dev"},
	"pdo_mysql":    []string{},
	"pdo_odbc":     []string{"unixodbc-dev", phpODBC},
	"pdo_pgsql":    []string{"libpq-dev"},
	"pdo_sqlite":   []string{"libsqlite3-dev"},
	"pgsql":        []string{"libpq-dev"},
	"pspell":       []string{"libpspell-dev"},
	"shmop":        []string{},
	"snmp":         []string{"libsnmp-dev"},
	"soap":         []string{"libxml2-dev"},
	"sockets":      []string{},
	"sysvmsg":      []string{},
	"sysvsem":      []string{},
	"sysvshm":      []string{},
	"tidy":         []string{"libtidy-dev"},
	"xsl":          []string{"libxslt1-dev"},
	"zip":          []string{"libzip-dev"},
}

var phpPeclExts = map[string][]string{
	"amqp":      []string{"librabbitmq-dev"},
	"apcu":      []string{},
	"igbinary":  []string{},
	"imagick":   []string{"libmagickwand-dev"},
	"lzf":       []string{},
	"mailparse": []string{},
	"maxminddb": []string{"libmaxminddb-dev"},
	"mcrypt":    []string{"libmcrypt-dev"},
	"memcached": []string{"libmemcached-dev"},
	"mongodb":   []string{},
	"msgpack":   []string{},
	"oauth":     []string{"libpcre3-dev"},
	"protobuf":  []string{},
	"psr":       []string{},
	"rdkafka":   []string{"librdkafka-dev"},
	"redis":     []string{},
	"solr":      []string{"libcurl4-openssl-dev,libxml2-dev"},
	"stomp":     []string{"libssl-dev"},
	"yaf":       []string{},
	"yaml":      []string{"libyaml-dev"},
}
