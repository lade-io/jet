package pack

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
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
		conf, core, make, pecl, pkgs := p.extensions()
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
				Args: make,
				List: true,
			})
		}

		pkgs = append(pkgs, []string{"git", "wget", "zip"}...)
		meta.Packages = append(meta.Packages, pkgs...)
		meta.Tools = append(meta.Tools, &Tool{
			Name:    "composer",
			Owner:   "composer",
			Files:   []string{"composer.json", "composer.lock"},
			Install: []string{"install --no-dev"},
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
	return "<7.4," + requires["php"], nil
}

func (p *PhpPack) extensions() ([]string, []string, []string, []string, []string) {
	requires := p.requires()
	var exts []string
	for require, _ := range requires {
		ext := strings.TrimPrefix(require, "ext-")
		if ext != require {
			exts = append(exts, ext)
		}
	}
	sort.Strings(exts)

	var conf, core, make, pecl, pkgs []string
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
			make = append(make, ext)
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
	return conf, core, make, pecl, pkgs
}

func (p *PhpPack) requires() map[string]string {
	b, err := fileRead(p.WorkDir, "composer.json")
	if err != nil {
		return nil
	}

	conf := map[string]interface{}{}
	if err = json.Unmarshal(b, &conf); err != nil {
		return nil
	}

	require, ok := conf["require"].(map[string]interface{})
	if !ok {
		return nil
	}

	requires := map[string]string{}
	for k, v := range require {
		val, ok := v.(string)
		if ok {
			requires[k] = val
		}
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

var phpLayout = regexp.MustCompile(`^require.*index\.php`)

const (
	phpImap = "--with-kerberos --with-imap-ssl"
	phpODBC = "--with-pdo-odbc=unixODBC,/usr"
)

var phpCoreExts = map[string][]string{
	"bz2":          []string{"libbz2-dev"},
	"curl":         []string{"libcurl4-openssl-dev"},
	"dba":          []string{},
	"enchant":      []string{"libenchant-dev"},
	"exif":         []string{},
	"fileinfo":     []string{},
	"ftp":          []string{"libssl-dev"},
	"gd":           []string{"libpng-dev"},
	"gettext":      []string{},
	"gmp":          []string{"libgmp-dev"},
	"imap":         []string{"libc-client-dev,libkrb5-dev", phpImap},
	"interbase":    []string{"firebird-dev"},
	"ldap":         []string{"libldap2-dev"},
	"mbstring":     []string{},
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
	"readline":     []string{"libedit-dev"},
	"recode":       []string{"librecode-dev"},
	"shmop":        []string{},
	"snmp":         []string{"libsnmp-dev"},
	"soap":         []string{"libxml2-dev"},
	"sockets":      []string{},
	"sysvmsg":      []string{},
	"sysvsem":      []string{},
	"sysvshm":      []string{},
	"tidy":         []string{"libtidy-dev"},
	"wddx":         []string{"libxml2-dev"},
	"xmlrpc":       []string{"libxml2-dev"},
	"xsl":          []string{"libxslt1-dev"},
	"zip":          []string{"libzip-dev"},
}

var phpPeclExts = map[string][]string{
	"amqp":      []string{"librabbitmq-dev"},
	"apcu":      []string{},
	"geoip":     []string{"libgeoip-dev", "beta"},
	"igbinary":  []string{},
	"imagick":   []string{"libmagickwand-dev"},
	"lzf":       []string{},
	"mcrypt":    []string{"libmcrypt-dev", "snapshot"},
	"memcached": []string{"libmemcached-dev"},
	"mongodb":   []string{},
	"msgpack":   []string{},
	"oauth":     []string{"libpcre3-dev"},
	"protobuf":  []string{},
	"rdkafka":   []string{"librdkafka-dev"},
	"redis":     []string{},
	"yaml":      []string{"libyaml-dev"},
}
