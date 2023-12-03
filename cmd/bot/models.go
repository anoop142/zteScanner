package main

import(
	"database/sql"
	"time"
	"context"
	"errors"
)


var(
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict = errors.New("edit conflict")
)



type Models struct{
	KnownDevices KnownDeviceModel
}

func NewModels(db *sql.DB) Models{
	return Models{
		KnownDevices : KnownDeviceModel{DB: db},
	}
}






type KnownDeviceModel struct{
	DB *sql.DB
}



func (k KnownDeviceModel) Insert(dev *KnownDevice)error{
	query := `
	INSERT INTO known_devices (name, mac, ignore)
	VALUES ($1, $2, $3)`

	args := []any{dev.Name, dev.Mac, dev.Ignore}

	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	_, err := k.DB.ExecContext(ctx, query, args...)
	if err != nil{
		return err
	}

	return nil

}


func (k KnownDeviceModel) GetIgnored()([]KnownDevice, error){
	var devices []KnownDevice
	query:=`SELECT mac, name, ignore
		from known_devices
		WHERE ignore=1`

	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	rows, err := k.DB.QueryContext(ctx, query)
	if err  != nil{
		return nil, err
	}
	defer rows.Close()


	for rows.Next(){
		var device KnownDevice
		err := rows.Scan(
			&device.Mac,
			&device.Name,
			&device.Ignore,
		)
		if err != nil{
			return nil, err
		}
		devices = append(devices, device)
	}
	if err:=rows.Err(); err != nil{
		return nil, err
	}
	if len(devices) == 0{
		return nil, errors.New("empty ignored devices")
	}
	return devices, nil
}


func (k KnownDeviceModel) GetKnown()([]KnownDevice, error){
	var devices []KnownDevice
	query:=`SELECT mac, name, ignore
		from known_devices
		WHERE ignore=0`

	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	rows, err := k.DB.QueryContext(ctx, query)
	if err  != nil{
		return nil, err
	}
	defer rows.Close()


	for rows.Next(){
		var device KnownDevice
		err := rows.Scan(
			&device.Mac,
			&device.Name,
			&device.Ignore,
		)
		if err != nil{
			return nil, err
		}
		devices = append(devices, device)
	}
	if err:=rows.Err(); err != nil{
		return nil, err
	}
	if len(devices) == 0{
		return nil, errors.New("empty known devices")
	}
	return devices, nil
}





func (k KnownDeviceModel)Get(mac string)(*KnownDevice, error){
	var device KnownDevice
	query:=`SELECT mac, name, ignore 
	from known_devices
	WHERE mac=$1`

	args:= []any{mac}

	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	err := k.DB.QueryRowContext(ctx, query, args...).Scan(
			&device.Mac,
			&device.Name,
			&device.Ignore,
		)

	if err != nil{
		switch{
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &device, nil
}



func(k KnownDeviceModel)Delete(mac string) error{
	query := `DELETE FROM known_devices
		  WHERE mac=$1`
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	result, err := k.DB.ExecContext(ctx, query, mac)
	if err != nil{
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil{
		return err
	}

	if rowsAffected == 0{
		return ErrRecordNotFound
	}

	return nil
}


	
